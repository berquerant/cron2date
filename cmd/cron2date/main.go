package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/berquerant/cron2date/internal/cron2date"
)

const usage = `cron2date - expand cron expression

Usage:

  cron2date [flags] CRONEXPR

Flags:`

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr, time.Now); err != nil {
		fmt.Fprintf(os.Stderr, "cron2date: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer, now func() time.Time) error {
	fs := flag.NewFlagSet("cron2date", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() {
		_, _ = fmt.Fprintln(stderr, usage)
		fs.PrintDefaults()
	}

	var (
		startString = fs.String("s", "", "start time (default current time)")
		endString   = fs.String("e", "", "end time")
		count       = fs.Int("c", -1, "firing times to display, negative value is ignored")
		duration    = fs.Duration(
			"d",
			time.Duration(-1),
			"set the end time to the start time + this duration, negative value is ignored",
		)
		timeFormat = fs.String("f", time.DateTime, "time layout")
	)

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	switch {
	case fs.NArg() == 0:
		return errors.New("CRONEXPR is required")
	case fs.NArg() > 1:
		return fmt.Errorf("unknown arguments: %v", fs.Args()[1:])
	}

	var (
		cronExpr = fs.Arg(0)
		loc      = time.Local
		start    = now().In(loc)
		end      time.Time
	)
	if *startString != "" {
		t, err := time.ParseInLocation(*timeFormat, *startString, loc)
		if err != nil {
			return fmt.Errorf("parse time %s: %w", *startString, err)
		}
		start = t
	}
	if *endString != "" {
		t, err := time.ParseInLocation(*timeFormat, *endString, loc)
		if err != nil {
			return fmt.Errorf("parse time %s: %w", *endString, err)
		}
		end = t
	}

	nexter, err := cron2date.NewNexter(cronExpr)
	if err != nil {
		return fmt.Errorf("parse cronexpr %s: %w", cronExpr, err)
	}

	seq, err := cron2date.NewIterator(nexter, cron2date.Config{
		Start:    start,
		Count:    *count,
		Duration: *duration,
		End:      end,
	})
	if err != nil {
		return fmt.Errorf("validate: %w", err)
	}

	for x := range seq {
		if _, err := fmt.Fprintf(stdout, "%s\n", x.Format(*timeFormat)); err != nil {
			return err
		}
	}
	return nil
}
