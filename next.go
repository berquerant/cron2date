package cron2date

import (
	"time"

	"github.com/robfig/cron"
)

type Nexter interface {
	// Next returns the next firing time.
	Next(time.Time) time.Time
}

type NextFunc func(time.Time) time.Time

func (f NextFunc) Next(t time.Time) time.Time {
	return f(t)
}

func NewNexter(spec string) (Nexter, error) {
	return cron.ParseStandard(spec)
}
