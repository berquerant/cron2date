package cron2date

import (
	"errors"
	"fmt"
	"time"
)

type Iterator interface {
	Next() bool
	Time() time.Time
}

//go:generate go run github.com/berquerant/goconfig -field "Start time.Time|Count int|Duration time.Duration|End time.Time" -option -configOption Option -output iter_config_generated.go

type IteratorImpl struct {
	config  *Config
	nexter  Nexter
	count   int
	current time.Time
}

func NewIteratorImpl(nexter Nexter, opt ...Option) (*IteratorImpl, error) {
	config := NewConfigBuilder().
		Count(-1).
		Duration(time.Duration(-1)).
		Build()
	config.Apply(opt...)
	if err := config.validate(); err != nil {
		return nil, err
	}
	return &IteratorImpl{
		config:  config,
		nexter:  nexter,
		count:   0,
		current: config.Start.Get(),
	}, nil
}

func (iter *IteratorImpl) Time() time.Time {
	return iter.current
}

func (iter *IteratorImpl) Next() bool {
	if iter.isEOI() {
		return false
	}
	iter.current = iter.nexter.Next(iter.current)
	iter.count++
	return !iter.isEOI()
}

func (iter *IteratorImpl) isEOI() bool {
	if iter.current == zeroTime {
		return true
	}
	if limit := iter.config.Count.Get(); limit >= 0 {
		if iter.count > limit {
			return true
		}
	}
	if duration := iter.config.Duration.Get(); duration >= 0 {
		if iter.current.After(iter.config.Start.Get().Add(duration)) {
			return true
		}
	}
	if end := iter.config.End.Get(); end.After(zeroTime) {
		if iter.current.After(end) {
			return true
		}
	}

	return false
}

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
