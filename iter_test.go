package cron2date_test

import (
	"testing"
	"time"

	"github.com/berquerant/cron2date"
	"github.com/stretchr/testify/assert"
)

func TestIterator(t *testing.T) {
	var (
		year2000 = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
		year3000 = time.Date(3000, 1, 1, 0, 0, 0, 0, time.UTC)
		genNext  = func(n int) []time.Time {
			result := make([]time.Time, n)
			for i := 0; i < n; i++ {
				result[i] = year2000.Add(time.Duration(i+1) * time.Hour)
			}
			return result
		}
		nexter = cron2date.NextFunc(func(t time.Time) time.Time {
			return t.Add(time.Hour)
		})
		mustNewIterator = func(opt ...cron2date.Option) cron2date.Iterator {
			iter, err := cron2date.NewIteratorImpl(nexter, opt...)
			if err != nil {
				t.Fatal(err)
			}
			return iter
		}
		start = cron2date.WithStart(year2000)
	)

	for _, tc := range []struct {
		title string
		iter  cron2date.Iterator
		want  []time.Time
	}{
		{
			title: "0 count",
			iter:  mustNewIterator(start, cron2date.WithCount(0)),
			want:  genNext(0),
		},
		{
			title: "3 count",
			iter:  mustNewIterator(start, cron2date.WithCount(3)),
			want:  genNext(3),
		},
		{
			title: "0 duration",
			iter:  mustNewIterator(start, cron2date.WithDuration(0)),
			want:  genNext(0),
		},
		{
			title: "3 duration",
			iter:  mustNewIterator(start, cron2date.WithDuration(3*time.Hour)),
			want:  genNext(3),
		},
		{
			title: "0 end",
			iter:  mustNewIterator(start, cron2date.WithEnd(year2000)),
			want:  genNext(0),
		},
		{
			title: "3 end",
			iter:  mustNewIterator(start, cron2date.WithEnd(year2000.Add(3*time.Hour))),
			want:  genNext(3),
		},
	} {
		tc := tc
		t.Run(tc.title, func(t *testing.T) {
			got := []time.Time{}
			for tc.iter.Next() {
				got = append(got, tc.iter.Time())
			}
			assert.Equal(t, tc.want, got)
		})
	}

	t.Run("Validate", func(t *testing.T) {
		for _, tc := range []struct {
			title string
			opt   []cron2date.Option
			isErr bool
		}{
			{
				title: "start is required",
				opt:   []cron2date.Option{cron2date.WithCount(1)},
				isErr: true,
			},
			{
				title: "end condition is required",
				opt:   []cron2date.Option{start},
				isErr: true,
			},
			{
				title: "end count",
				opt:   []cron2date.Option{start, cron2date.WithCount(1)},
			},
			{
				title: "end duration",
				opt:   []cron2date.Option{start, cron2date.WithDuration(time.Hour)},
			},
			{
				title: "end time",
				opt:   []cron2date.Option{start, cron2date.WithEnd(year3000)},
			},
		} {
			tc := tc
			t.Run(tc.title, func(t *testing.T) {
				_, err := cron2date.NewIteratorImpl(nexter, tc.opt...)
				assert.Equal(t, tc.isErr, err != nil, "got %v", err)
			})
		}
	})
}
