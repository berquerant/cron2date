package cron2date

import (
	"errors"
	"fmt"
	"iter"
	"runtime/debug"
	"time"
)

// DefaultVersion is the fallback version string when build info is unavailable.
const DefaultVersion = "(devel)"

// BuildVersion returns the main module version from Go build info, or DefaultVersion.
func BuildVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" {
		return info.Main.Version
	}
	return DefaultVersion
}

// ResolveLocation loads an IANA timezone by name, defaulting to time.Local when empty.
func ResolveLocation(name string) (*time.Location, error) {
	if name == "" {
		return time.Local, nil
	}
	loc, err := time.LoadLocation(name)
	if err != nil {
		return nil, fmt.Errorf("load timezone %s: %w", name, err)
	}
	return loc, nil
}

// ResolveFormat returns format, defaulting to time.DateTime when empty.
func ResolveFormat(format string) string {
	if format == "" {
		return time.DateTime
	}
	return format
}

// ParseTimeOrDefault parses s in layout and loc, returning fallback when s is empty.
func ParseTimeOrDefault(s, layout string, loc *time.Location, fallback time.Time) (time.Time, error) {
	if s == "" {
		return fallback, nil
	}
	t, err := time.ParseInLocation(layout, s, loc)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse time %s: %w", s, err)
	}
	return t, nil
}

// ParseNexter parses spec as a 5-field standard cron expression or a 6-field seconds expression.
func ParseNexter(spec string, seconds bool) (Nexter, error) {
	if seconds {
		return NewNexterWithSeconds(spec)
	}
	return NewNexter(spec)
}

// ExpandParams holds common expansion parameters shared by CLI and MCP adapters.
type ExpandParams struct {
	Expr     string
	Start    string
	End      string
	Count    int           // negative value is ignored
	Duration time.Duration // negative value is ignored
	Format   string
	Timezone string
	Seconds  bool
}

// PreparedExpand holds the validated iterator and resolved formatting/location settings.
type PreparedExpand struct {
	Seq      iter.Seq[time.Time]
	Format   string
	Location *time.Location
}

// PrepareExpand resolves timezone, layout, timestamps, and cron expression into a ready-to-run iterator.
func PrepareExpand(params ExpandParams, now func() time.Time) (PreparedExpand, error) {
	if params.Expr == "" {
		return PreparedExpand{}, errors.New("expr is required")
	}
	if now == nil {
		now = time.Now
	}

	loc, err := ResolveLocation(params.Timezone)
	if err != nil {
		return PreparedExpand{}, err
	}
	layout := ResolveFormat(params.Format)

	start, err := ParseTimeOrDefault(params.Start, layout, loc, now().In(loc))
	if err != nil {
		return PreparedExpand{}, err
	}
	end, err := ParseTimeOrDefault(params.End, layout, loc, time.Time{})
	if err != nil {
		return PreparedExpand{}, err
	}

	nexter, err := ParseNexter(params.Expr, params.Seconds)
	if err != nil {
		return PreparedExpand{}, fmt.Errorf("parse cronexpr %s: %w", params.Expr, err)
	}

	seq, err := NewIterator(nexter, Config{
		Start:    start,
		Count:    params.Count,
		Duration: params.Duration,
		End:      end,
	})
	if err != nil {
		return PreparedExpand{}, fmt.Errorf("validate: %w", err)
	}

	return PreparedExpand{
		Seq:      seq,
		Format:   layout,
		Location: loc,
	}, nil
}
