package cron2date_test

import (
	"slices"
	"testing"
	"time"

	"github.com/berquerant/cron2date/internal/cron2date"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrepareExpand(t *testing.T) {
	fixedNow := func() time.Time {
		return time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	}

	t.Run("build version is non-empty", func(t *testing.T) {
		assert.NotEmpty(t, cron2date.BuildVersion())
	})

	t.Run("success with defaults and options", func(t *testing.T) {
		prepared, err := cron2date.PrepareExpand(cron2date.ExpandParams{
			Expr:     "0 * * * *",
			Count:    2,
			Duration: -1,
			Timezone: "UTC",
		}, fixedNow)
		require.NoError(t, err)
		assert.Equal(t, time.DateTime, prepared.Format)
		assert.Equal(t, "UTC", prepared.Location.String())

		got := slices.Collect(prepared.Seq)
		assert.Equal(t, []time.Time{
			time.Date(2026, 6, 1, 11, 0, 0, 0, time.UTC),
			time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC),
		}, got)
	})

	t.Run("seconds and explicit start/end", func(t *testing.T) {
		prepared, err := cron2date.PrepareExpand(cron2date.ExpandParams{
			Expr:     "*/30 * * * * *",
			Start:    "2026-06-01 10:00:00",
			End:      "2026-06-01 10:01:00",
			Count:    -1,
			Duration: -1,
			Timezone: "UTC",
			Seconds:  true,
		}, nil)
		require.NoError(t, err)
		got := slices.Collect(prepared.Seq)
		assert.Len(t, got, 2)
	})

	t.Run("error cases", func(t *testing.T) {
		for _, tc := range []struct {
			title  string
			params cron2date.ExpandParams
		}{
			{
				title:  "empty expr",
				params: cron2date.ExpandParams{Count: 1, Duration: -1},
			},
			{
				title:  "invalid timezone",
				params: cron2date.ExpandParams{Expr: "0 * * * *", Count: 1, Duration: -1, Timezone: "Invalid/Zone"},
			},
			{
				title:  "invalid start",
				params: cron2date.ExpandParams{Expr: "0 * * * *", Count: 1, Duration: -1, Start: "bad"},
			},
			{
				title:  "invalid end",
				params: cron2date.ExpandParams{Expr: "0 * * * *", Count: -1, Duration: -1, End: "bad"},
			},
			{
				title:  "invalid expr",
				params: cron2date.ExpandParams{Expr: "invalid", Count: 1, Duration: -1},
			},
			{
				title:  "missing stop condition",
				params: cron2date.ExpandParams{Expr: "0 * * * *", Count: -1, Duration: -1},
			},
		} {
			t.Run(tc.title, func(t *testing.T) {
				_, err := cron2date.PrepareExpand(tc.params, fixedNow)
				assert.Error(t, err)
			})
		}
	})
}
