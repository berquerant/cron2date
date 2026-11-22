package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/berquerant/cron2date/internal/cron2date"
)

const usage = `cron2date - expand cron expression

Usage:

  cron2date [flags] CRONEXPR

Flags:`

func main() {
	os.Exit(runMain())
}

func runMain() int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, os.Args[1:], os.Stdout, os.Stderr, time.Now); err != nil {
		if errors.Is(err, context.Canceled) {
			return 0
		}
		fmt.Fprintf(os.Stderr, "cron2date: %v\n", err)
		return 1
	}
	return 0
}

type cliOptions struct {
	startString string
	endString   string
	count       int
	duration    time.Duration
	timeFormat  string
	timezone    string
	seconds     bool
	version     bool
	cronExpr    string
}

func parseFlags(args []string, stderr io.Writer) (cliOptions, bool, error) {
	fs := flag.NewFlagSet("cron2date", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() {
		_, _ = fmt.Fprintln(stderr, usage)
		fs.PrintDefaults()
	}

	var opts cliOptions
	fs.StringVar(&opts.startString, "s", "", "start time (default current time)")
	fs.StringVar(&opts.endString, "e", "", "end time")
	fs.IntVar(&opts.count, "c", -1, "firing times to display, negative value is ignored")
	fs.DurationVar(
		&opts.duration,
		"d",
		time.Duration(-1),
		"set the end time to the start time + this duration, negative value is ignored",
	)
	fs.StringVar(&opts.timeFormat, "f", time.DateTime, "time layout")
	fs.StringVar(&opts.timezone, "tz", "", "timezone name (e.g. UTC, Asia/Tokyo; default local)")
	fs.BoolVar(&opts.seconds, "seconds", false, "parse 6-field cron expression with leading seconds")
	fs.BoolVar(&opts.version, "version", false, "print version and exit")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return opts, true, nil
		}
		return opts, false, err
	}
	if opts.version {
		return opts, false, nil
	}

	switch {
	case fs.NArg() == 0:
		return opts, false, errors.New("CRONEXPR is required")
	case fs.NArg() > 1:
		return opts, false, fmt.Errorf("unknown arguments: %v", fs.Args()[1:])
	}
	opts.cronExpr = fs.Arg(0)
	return opts, false, nil
}

func buildVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" {
		return info.Main.Version
	}
	return "(devel)"
}

func resolveLocation(name string) (*time.Location, error) {
	if name == "" {
		return time.Local, nil
	}
	loc, err := time.LoadLocation(name)
	if err != nil {
		return nil, fmt.Errorf("load timezone %s: %w", name, err)
	}
	return loc, nil
}

func parseTimeOrDefault(s, layout string, loc *time.Location, fallback time.Time) (time.Time, error) {
	if s == "" {
		return fallback, nil
	}
	t, err := time.ParseInLocation(layout, s, loc)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse time %s: %w", s, err)
	}
	return t, nil
}

func newNexter(spec string, withSeconds bool) (cron2date.Nexter, error) {
	if withSeconds {
		return cron2date.NewNexterWithSeconds(spec)
	}
	return cron2date.NewNexter(spec)
}

func run(ctx context.Context, args []string, stdout, stderr io.Writer, now func() time.Time) error {
	opts, helped, err := parseFlags(args, stderr)
	if err != nil || helped {
		return err
	}
	if opts.version {
		_, err := fmt.Fprintln(stdout, buildVersion())
		return err
	}

	loc, err := resolveLocation(opts.timezone)
	if err != nil {
		return err
	}

	start, err := parseTimeOrDefault(opts.startString, opts.timeFormat, loc, now().In(loc))
	if err != nil {
		return err
	}
	end, err := parseTimeOrDefault(opts.endString, opts.timeFormat, loc, time.Time{})
	if err != nil {
		return err
	}

	nexter, err := newNexter(opts.cronExpr, opts.seconds)
	if err != nil {
		return fmt.Errorf("parse cronexpr %s: %w", opts.cronExpr, err)
	}

	seq, err := cron2date.NewIterator(nexter, cron2date.Config{
		Start:    start,
		Count:    opts.count,
		Duration: opts.duration,
		End:      end,
	})
	if err != nil {
		return fmt.Errorf("validate: %w", err)
	}

	w := bufio.NewWriter(stdout)
	for x := range seq {
		if err := ctx.Err(); err != nil {
			_ = w.Flush()
			return err
		}
		if _, err := fmt.Fprintf(w, "%s\n", x.Format(opts.timeFormat)); err != nil {
			return err
		}
	}
	return w.Flush()
}
