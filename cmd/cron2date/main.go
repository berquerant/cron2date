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
	"syscall"
	"time"

	"github.com/berquerant/cron2date/internal/cron2date"
	cronmcp "github.com/berquerant/cron2date/internal/mcp"
)

const usage = `cron2date - expand cron expression

Usage:

  cron2date [flags] CRONEXPR
  cron2date -mcp

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
	params  cron2date.ExpandParams
	version bool
	mcpMode bool
}

func parseFlags(args []string, stderr io.Writer) (cliOptions, bool, error) {
	fs := flag.NewFlagSet("cron2date", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() {
		_, _ = fmt.Fprintln(stderr, usage)
		fs.PrintDefaults()
	}

	var opts cliOptions
	fs.StringVar(&opts.params.Start, "s", "", "start time (default current time)")
	fs.StringVar(&opts.params.End, "e", "", "end time")
	fs.IntVar(&opts.params.Count, "c", -1, "firing times to display, negative value is ignored")
	fs.DurationVar(
		&opts.params.Duration,
		"d",
		time.Duration(-1),
		"set the end time to the start time + this duration, negative value is ignored",
	)
	fs.StringVar(&opts.params.Format, "f", time.DateTime, "time layout")
	fs.StringVar(&opts.params.Timezone, "tz", "", "timezone name (e.g. UTC, Asia/Tokyo; default local)")
	fs.BoolVar(&opts.params.Seconds, "seconds", false, "parse 6-field cron expression with leading seconds")
	fs.BoolVar(&opts.mcpMode, "mcp", false, "run as an MCP server over stdio")
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
	if opts.mcpMode {
		if fs.NArg() > 0 {
			return opts, false, fmt.Errorf("unknown arguments: %v", fs.Args())
		}
		return opts, false, nil
	}

	switch {
	case fs.NArg() == 0:
		return opts, false, errors.New("CRONEXPR is required")
	case fs.NArg() > 1:
		return opts, false, fmt.Errorf("unknown arguments: %v", fs.Args()[1:])
	}
	opts.params.Expr = fs.Arg(0)
	return opts, false, nil
}

func run(ctx context.Context, args []string, stdout, stderr io.Writer, now func() time.Time) error {
	opts, helped, err := parseFlags(args, stderr)
	if err != nil || helped {
		return err
	}
	if opts.version {
		_, err := fmt.Fprintln(stdout, cron2date.BuildVersion())
		return err
	}
	if opts.mcpMode {
		return cronmcp.RunStdio(ctx, cronmcp.Options{
			Version: cron2date.BuildVersion(),
			Now:     now,
		})
	}

	prepared, err := cron2date.PrepareExpand(opts.params, now)
	if err != nil {
		return err
	}

	w := bufio.NewWriter(stdout)
	for x := range prepared.Seq {
		if err := ctx.Err(); err != nil {
			_ = w.Flush()
			return err
		}
		if _, err := fmt.Fprintf(w, "%s\n", x.In(prepared.Location).Format(prepared.Format)); err != nil {
			return err
		}
	}
	return w.Flush()
}
