package cron2date

import (
	"errors"
	"fmt"
	"iter"
	"time"
)

func NewIterator(nexter Nexter, opt ...Option) (iter.Seq[time.Time], error) {
	config := NewConfigBuilder().
		Count(-1).
		Duration(time.Duration(-1)).
		Build()
	config.Apply(opt...)
	if err := config.validate(); err != nil {
		return nil, err
	}

	var (
		current = config.Start.Get()
		count   int
		isEOI   = func() bool {
			if current == zeroTime {
				return true
			}
			if limit := config.Count.Get(); limit >= 0 && count > limit {
				return true
			}
			if duration := config.Duration.Get(); duration >= 0 && current.After(config.Start.Get().Add(duration)) {
				return true
			}
			if end := config.End.Get(); end.After(zeroTime) && current.After(end) {
				return true
			}
			return false
		}
	)

	return func(yield func(time.Time) bool) {
		for {
			current = nexter.Next(current)
			count++
			if isEOI() || !yield(current) {
				return
			}
		}
	}, nil
}

//go:generate go run github.com/berquerant/goconfig -field "Start time.Time|Count int|Duration time.Duration|End time.Time" -option -configOption Option -output iter_config_generated.go

var (
	zeroTime          time.Time
	ErrIteratorConfig = errors.New("IteratorConfig")
)

func (c *Config) validate() error {
	if !c.Start.IsModified() {
		return fmt.Errorf("%w, start is required", ErrIteratorConfig)
	}
	if !(c.Count.Get() >= 0 || c.Duration.Get() >= 0 || c.End.Get().After(zeroTime)) {
		return fmt.Errorf("%w, either count, duration or end is required", ErrIteratorConfig)
	}
	return nil
}
