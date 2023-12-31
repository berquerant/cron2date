package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/berquerant/cron2date"
)

const usage = `date2cron - convert time to cronexpr

Usage:

  date2cron [flags] [TIME]

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
		timeFormat = flag.String("f", time.DateTime, "time layout")
	)

	log.SetFlags(0)
	log.SetPrefix("date2cron: ")

	flag.Usage = Usage
	flag.Parse()

	var timeString = ""
	switch {
	case flag.NArg() == 1:
		timeString = flag.Arg(0)
	case flag.NArg() > 1:
		log.Fatalf("unknown arguments: %v", flag.Args()[1:])
	}

	var targetTime = time.Now()
	if timeString != "" {
		t, err := time.Parse(*timeFormat, timeString)
		fail(err, "invalid time")
		targetTime = t
	}

	fmt.Println(cron2date.Date2Cron(targetTime))
}
