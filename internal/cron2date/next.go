package cron2date

import (
	"time"

	"github.com/robfig/cron/v3"
)

// Nexter computes the next firing time after a given timestamp.
type Nexter interface {
	// Next returns the next firing time strictly after t, or the zero Time if none exists.
	Next(t time.Time) time.Time
}

// NextFunc adapts a plain function to the Nexter interface.
type NextFunc func(time.Time) time.Time

func (f NextFunc) Next(t time.Time) time.Time {
	return f(t)
}

// NewNexter parses a standard 5-field cron expression or descriptor.
func NewNexter(spec string) (Nexter, error) {
	return cron.ParseStandard(spec)
}
