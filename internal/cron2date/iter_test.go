package cron2date_test

import (
	"iter"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/berquerant/cron2date/internal/cron2date"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIterator(t *testing.T) {
	var (
		year2000 = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
		year3000 = time.Date(3000, 1, 1, 0, 0, 0, 0, time.UTC)
		genNext  = func(n int) []time.Time {
			result := make([]time.Time, n)
			for i := range n {
				result[i] = year2000.Add(time.Duration(i+1) * time.Hour)
			}
			return result
		}
		nexter = cron2date.NextFunc(func(t time.Time) time.Time {
			return t.Add(time.Hour)
		})
		mustNewIterator = func(cfg cron2date.Config) iter.Seq[time.Time] {
			t.Helper()
			seq, err := cron2date.NewIterator(nexter, cfg)
			require.NoError(t, err)
			return seq
		}
		cfgWith = func(mod func(*cron2date.Config)) cron2date.Config {
			cfg := cron2date.DefaultConfig()
			cfg.Start = year2000
			mod(&cfg)
			return cfg
		}
	)

	for _, tc := range []struct {
		title string
		iter  iter.Seq[time.Time]
		want  []time.Time
	}{
		{
			title: "0 count",
			iter:  mustNewIterator(cfgWith(func(c *cron2date.Config) { c.Count = 0 })),
			want:  genNext(0),
		},
		{
			title: "3 count",
			iter:  mustNewIterator(cfgWith(func(c *cron2date.Config) { c.Count = 3 })),
			want:  genNext(3),
		},
		{
			title: "0 duration",
			iter:  mustNewIterator(cfgWith(func(c *cron2date.Config) { c.Duration = 0 })),
			want:  genNext(0),
		},
		{
			title: "3 duration",
			iter:  mustNewIterator(cfgWith(func(c *cron2date.Config) { c.Duration = 3 * time.Hour })),
			want:  genNext(3),
		},
		{
			title: "0 end",
			iter:  mustNewIterator(cfgWith(func(c *cron2date.Config) { c.End = year2000 })),
			want:  genNext(0),
		},
		{
			title: "3 end",
			iter:  mustNewIterator(cfgWith(func(c *cron2date.Config) { c.End = year2000.Add(3 * time.Hour) })),
			want:  genNext(3),
		},
		{
			title: "combined count and duration stops at earlier limit",
			iter: mustNewIterator(cfgWith(func(c *cron2date.Config) {
				c.Count = 10
				c.Duration = 2 * time.Hour
			})),
			want: genNext(2),
		},
	} {
		t.Run(tc.title, func(t *testing.T) {
			got := slices.Collect(tc.iter)
			if len(tc.want) == 0 {
				assert.Empty(t, got)
				return
			}
			assert.Equal(t, tc.want, got)
		})
	}

	t.Run("reusable and concurrent safe", func(t *testing.T) {
		seq := mustNewIterator(cfgWith(func(c *cron2date.Config) { c.Count = 3 }))
		want := genNext(3)

		// First pass with early break
		var partial []time.Time
		for v := range seq {
			partial = append(partial, v)
			break
		}
		assert.Equal(t, want[:1], partial)

		// Second pass must start from the beginning
		assert.Equal(t, want, slices.Collect(seq))

		// Concurrent iteration must be race-free
		var wg sync.WaitGroup
		for range 8 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				assert.Equal(t, want, slices.Collect(seq))
			}()
		}
		wg.Wait()
	})

	t.Run("count 0 does not call nexter", func(t *testing.T) {
		called := 0
		countingNexter := cron2date.NextFunc(func(t time.Time) time.Time {
			called++
			return t.Add(time.Hour)
		})
		seq, err := cron2date.NewIterator(countingNexter, cfgWith(func(c *cron2date.Config) { c.Count = 0 }))
		require.NoError(t, err)
		assert.Empty(t, slices.Collect(seq))
		assert.Equal(t, 0, called)
	})

	t.Run("stops on zero time or non-monotonic nexter", func(t *testing.T) {
		t.Run("zero time in non-UTC location", func(t *testing.T) {
			jst := time.FixedZone("JST", 9*3600)
			zeroInJST := time.Time{}.In(jst)
			zNexter := cron2date.NextFunc(func(time.Time) time.Time { return zeroInJST })
			seq, err := cron2date.NewIterator(zNexter, cfgWith(func(c *cron2date.Config) { c.Count = 5 }))
			require.NoError(t, err)
			assert.Empty(t, slices.Collect(seq))
		})

		t.Run("non-monotonic nexter terminates", func(t *testing.T) {
			stuckNexter := cron2date.NextFunc(func(t time.Time) time.Time { return t })
			seq, err := cron2date.NewIterator(stuckNexter, cfgWith(func(c *cron2date.Config) { c.Duration = time.Hour }))
			require.NoError(t, err)
			assert.Empty(t, slices.Collect(seq))
		})
	})

	t.Run("Validate", func(t *testing.T) {
		for _, tc := range []struct {
			title  string
			nexter cron2date.Nexter
			cfg    cron2date.Config
			isErr  bool
		}{
			{
				title:  "nexter is required",
				nexter: nil,
				cfg:    cfgWith(func(c *cron2date.Config) { c.Count = 1 }),
				isErr:  true,
			},
			{
				title:  "start is required",
				nexter: nexter,
				cfg:    cron2date.Config{Count: 1, Duration: -1},
				isErr:  true,
			},
			{
				title:  "end condition is required",
				nexter: nexter,
				cfg:    cfgWith(func(*cron2date.Config) {}),
				isErr:  true,
			},
			{
				title:  "end count",
				nexter: nexter,
				cfg:    cfgWith(func(c *cron2date.Config) { c.Count = 1 }),
			},
			{
				title:  "end duration",
				nexter: nexter,
				cfg:    cfgWith(func(c *cron2date.Config) { c.Duration = time.Hour }),
			},
			{
				title:  "end time",
				nexter: nexter,
				cfg:    cfgWith(func(c *cron2date.Config) { c.End = year3000 }),
			},
		} {
			t.Run(tc.title, func(t *testing.T) {
				_, err := cron2date.NewIterator(tc.nexter, tc.cfg)
				if tc.isErr {
					assert.ErrorIs(t, err, cron2date.ErrIteratorConfig)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})
}
