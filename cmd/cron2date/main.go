package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/berquerant/cron2date"
)

const usage = `cron2date - expand cron expression

Usage:

  cron2date [flags] CRONEXPR

Flags:`

func Usage() {
	fmt.Fprintln(os.Stderr, usage)
	flag.PrintDefaults()
}

func fail(err error, msg string, v ...any) {
	if err != nil {
		log.Fatalf("%s, %v", fmt.Sprintf(msg, v...), err)
	}
}

func main() {
	var (
		startString = flag.String("s", "", "start time (default current time)")
		endString   = flag.String("e", "", "end time")
		count       = flag.Int("c", -1, "firing times to display, negative value is ignored")
		duration    = flag.Duration("d", time.Duration(-1), "set the end time to the start time + this duration, negative value is ignored")
		timeFormat  = flag.String("f", time.DateTime, "time layout")
	)

	log.SetFlags(0)
	log.SetPrefix("cron2date: ")

	flag.Usage = Usage
	flag.Parse()

	switch {
	case flag.NArg() == 0:
		log.Fatal("CRONEXPR is required")
	case flag.NArg() > 1:
		log.Fatalf("unknown arguments: %v", flag.Args()[1:])
	}

	var (
		cronExpr  = flag.Arg(0)
		start     = time.Now()
		end       time.Time
		parseTime = func(s string) time.Time {
			t, err := time.Parse(*timeFormat, s)
			fail(err, "parse time %s", s)
			return t
		}
	)
	if *startString != "" {
		start = parseTime(*startString)
	}
	if *endString != "" {
		end = parseTime(*endString)
	}

	nexter, err := cron2date.NewNexter(cronExpr)
	fail(err, "parse cronexpr %s", cronExpr)

	iter, err := cron2date.NewIterator(
		nexter,
		cron2date.WithStart(start),
		cron2date.WithCount(*count),
		cron2date.WithDuration(*duration),
		cron2date.WithEnd(end),
	)
	fail(err, "validate")

	for x := range iter {
		fmt.Printf("%s\n", x.Format(*timeFormat))
	}
}
