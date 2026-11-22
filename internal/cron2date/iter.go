package cron2date

import (
	"errors"
	"fmt"
	"iter"
	"time"
)

// ErrIteratorConfig indicates an invalid iterator configuration.
var ErrIteratorConfig = errors.New("IteratorConfig")

// Config configures the firing-time iterator.
// Start is required, and at least one stop condition must be specified:
// Count >= 0, Duration >= 0, or a non-zero End.
// By default firing times are strictly after Start; set Inclusive to include Start when it matches.
type Config struct {
	Start     time.Time
	Count     int           // negative value is ignored
	Duration  time.Duration // negative value is ignored
	End       time.Time
	Inclusive bool
}

// DefaultConfig returns a Config with unset Count and Duration.
func DefaultConfig() Config {
	return Config{
		Count:    -1,
		Duration: -1,
	}
}

func (c Config) validate() error {
	if c.Start.IsZero() {
		return fmt.Errorf("%w: start is required", ErrIteratorConfig)
	}
	if c.Count < 0 && c.Duration < 0 && c.End.IsZero() {
		return fmt.Errorf("%w: either count, duration or end is required", ErrIteratorConfig)
	}
	return nil
}

func (c Config) deadline() (time.Time, bool) {
	var d time.Time
	has := false
	if c.Duration >= 0 {
		d = c.Start.Add(c.Duration)
		has = true
	}
	if !c.End.IsZero() && (!has || c.End.Before(d)) {
		d = c.End
		has = true
	}
	return d, has
}

func matchInclusiveStart(nexter Nexter, start, deadline time.Time, hasDeadline bool) (time.Time, bool) {
	candidate := nexter.Next(start.Add(-time.Nanosecond))
	if candidate.IsZero() || !candidate.Equal(start) {
		return time.Time{}, false
	}
	if hasDeadline && candidate.After(deadline) {
		return time.Time{}, false
	}
	return candidate, true
}

type iterator struct {
	nexter      Nexter
	cfg         Config
	deadline    time.Time
	hasDeadline bool
}

func (it iterator) run(yield func(time.Time) bool) {
	current := it.cfg.Start
	count := 0
	if it.cfg.Inclusive && it.underCountLimit(count) {
		if first, ok := matchInclusiveStart(it.nexter, current, it.deadline, it.hasDeadline); ok {
			if !yield(first) {
				return
			}
			count++
		}
	}
	for ; it.underCountLimit(count); count++ {
		next := it.nexter.Next(current)
		if next.IsZero() || !next.After(current) {
			return
		}
		if it.hasDeadline && next.After(it.deadline) {
			return
		}
		current = next
		if !yield(current) {
			return
		}
	}
}

func (it iterator) underCountLimit(count int) bool {
	return it.cfg.Count < 0 || count < it.cfg.Count
}

// NewIterator returns a reusable, stateless iterator yielding firing times
// from cfg.Start until the configured stop condition is reached.
func NewIterator(nexter Nexter, cfg Config) (iter.Seq[time.Time], error) {
	if nexter == nil {
		return nil, fmt.Errorf("%w: nexter is required", ErrIteratorConfig)
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	deadline, hasDeadline := cfg.deadline()
	it := iterator{
		nexter:      nexter,
		cfg:         cfg,
		deadline:    deadline,
		hasDeadline: hasDeadline,
	}
	return it.run, nil
}
