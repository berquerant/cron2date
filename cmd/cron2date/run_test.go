package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	ctx := context.Background()
	fixedNow := func() time.Time {
		return time.Date(2026, 6, 1, 10, 0, 0, 0, time.Local)
	}

	t.Run("help", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		err := run(ctx, []string{"-h"}, &stdout, &stderr, fixedNow)
		require.NoError(t, err)
		assert.Empty(t, stdout.String())
		assert.Contains(t, stderr.String(), "cron2date - expand cron expression")
	})

	t.Run("version", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		err := run(ctx, []string{"-version"}, &stdout, &stderr, fixedNow)
		require.NoError(t, err)
		assert.NotEmpty(t, strings.TrimSpace(stdout.String()))
		assert.Empty(t, stderr.String())
	})

	t.Run("canceled context stops iteration", func(t *testing.T) {
		canceledCtx, cancel := context.WithCancel(ctx)
		cancel()
		var stdout, stderr bytes.Buffer
		err := run(canceledCtx, []string{"-c", "10", "0 * * * *"}, &stdout, &stderr, fixedNow)
		assert.ErrorIs(t, err, context.Canceled)
		assert.Empty(t, stdout.String())
	})

	t.Run("success cases", func(t *testing.T) {
		for _, tc := range []struct {
			title string
			args  []string
			want  string
		}{
			{
				title: "default start with count",
				args:  []string{"-c", "2", "0 * * * *"},
				want:  "2026-06-01 11:00:00\n2026-06-01 12:00:00\n",
			},
			{
				title: "default start with end uses same local timezone",
				args:  []string{"-e", "2026-06-01 12:00:00", "0 * * * *"},
				want:  "2026-06-01 11:00:00\n2026-06-01 12:00:00\n",
			},
			{
				title: "explicit start and duration",
				args:  []string{"-s", "2026-06-01 10:00:00", "-d", "2h30m", "0 * * * *"},
				want:  "2026-06-01 11:00:00\n2026-06-01 12:00:00\n",
			},
			{
				title: "explicit timezone Asia/Tokyo with RFC3339",
				args: []string{
					"-tz", "Asia/Tokyo",
					"-f", time.RFC3339,
					"-s", "2026-06-01T10:00:00+09:00",
					"-c", "2",
					"0 * * * *",
				},
				want: "2026-06-01T11:00:00+09:00\n2026-06-01T12:00:00+09:00\n",
			},
			{
				title: "6-field seconds expression",
				args: []string{
					"-seconds",
					"-s", "2026-06-01 10:00:00",
					"-c", "2",
					"*/15 * * * * *",
				},
				want: "2026-06-01 10:00:15\n2026-06-01 10:00:30\n",
			},
			{
				title: "count 0 outputs empty",
				args:  []string{"-c", "0", "0 * * * *"},
				want:  "",
			},
		} {
			t.Run(tc.title, func(t *testing.T) {
				var stdout, stderr bytes.Buffer
				err := run(ctx, tc.args, &stdout, &stderr, fixedNow)
				require.NoError(t, err)
				assert.Equal(t, tc.want, stdout.String())
				assert.Empty(t, stderr.String())
			})
		}
	})

	t.Run("error cases", func(t *testing.T) {
		for _, tc := range []struct {
			title   string
			args    []string
			wantErr string
		}{
			{
				title:   "missing cronexpr",
				args:    []string{"-c", "3"},
				wantErr: "CRONEXPR is required",
			},
			{
				title:   "extra arguments",
				args:    []string{"-c", "3", "0 * * * *", "extra"},
				wantErr: "unknown arguments",
			},
			{
				title:   "unknown flag",
				args:    []string{"-unknown", "0 * * * *"},
				wantErr: "flag provided but not defined",
			},
			{
				title:   "invalid timezone",
				args:    []string{"-tz", "Invalid/Zone", "-c", "1", "0 * * * *"},
				wantErr: "load timezone Invalid/Zone",
			},
			{
				title:   "invalid start time",
				args:    []string{"-s", "not-a-time", "-c", "1", "0 * * * *"},
				wantErr: "parse time not-a-time",
			},
			{
				title:   "invalid end time",
				args:    []string{"-e", "not-a-time", "0 * * * *"},
				wantErr: "parse time not-a-time",
			},
			{
				title:   "invalid cron expression",
				args:    []string{"-c", "1", "invalid"},
				wantErr: "parse cronexpr invalid",
			},
			{
				title:   "5-field expression with -seconds fails",
				args:    []string{"-seconds", "-c", "1", "0 * * * *"},
				wantErr: "parse cronexpr",
			},
			{
				title:   "missing stop condition",
				args:    []string{"0 * * * *"},
				wantErr: "validate",
			},
		} {
			t.Run(tc.title, func(t *testing.T) {
				var stdout, stderr bytes.Buffer
				err := run(ctx, tc.args, &stdout, &stderr, fixedNow)
				require.Error(t, err)
				assert.True(t, strings.Contains(err.Error(), tc.wantErr), "got %v, want substring %q", err, tc.wantErr)
			})
		}
	})
}
