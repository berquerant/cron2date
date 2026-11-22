package cron2date_test

import (
	"slices"
	"testing"
	"time"

	"github.com/berquerant/cron2date/internal/cron2date"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewNexter(t *testing.T) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	t.Run("valid expressions", func(t *testing.T) {
		for _, tc := range []struct {
			title string
			spec  string
			want  time.Time
		}{
			{
				title: "hourly standard",
				spec:  "0 * * * *",
				want:  time.Date(2026, 1, 1, 1, 0, 0, 0, time.UTC),
			},
			{
				title: "specific minute and hour",
				spec:  "30 9 * * *",
				want:  time.Date(2026, 1, 1, 9, 30, 0, 0, time.UTC),
			},
			{
				title: "descriptor @daily",
				spec:  "@daily",
				want:  time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
			},
			{
				title: "descriptor @every",
				spec:  "@every 15m",
				want:  time.Date(2026, 1, 1, 0, 15, 0, 0, time.UTC),
			},
			{
				title: "CRON_TZ prefix",
				spec:  "CRON_TZ=Asia/Tokyo 0 10 * * *",
				// 2026-01-01 00:00 UTC is 09:00 JST, so next 10:00 JST is 01:00 UTC.
				want: time.Date(2026, 1, 1, 1, 0, 0, 0, time.UTC),
			},
		} {
			t.Run(tc.title, func(t *testing.T) {
				nexter, err := cron2date.NewNexter(tc.spec)
				require.NoError(t, err)
				got := nexter.Next(base)
				assert.True(t, tc.want.Equal(got), "want %v, got %v", tc.want, got)
			})
		}
	})

	t.Run("unmatchable expression terminates cleanly", func(t *testing.T) {
		nexter, err := cron2date.NewNexter("0 0 31 2 *")
		require.NoError(t, err)
		assert.True(t, nexter.Next(base).IsZero())

		cfg := cron2date.DefaultConfig()
		cfg.Start = base
		cfg.Count = 5
		seq, err := cron2date.NewIterator(nexter, cfg)
		require.NoError(t, err)
		assert.Empty(t, slices.Collect(seq))
	})

	t.Run("invalid expressions", func(t *testing.T) {
		for _, spec := range []string{
			"",
			"invalid",
			"* * * *",
			"* * * * * *",
			"60 * * * *",
		} {
			t.Run(spec, func(t *testing.T) {
				_, err := cron2date.NewNexter(spec)
				assert.Error(t, err)
			})
		}
	})
}
