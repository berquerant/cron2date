package mcp_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	cronmcp "github.com/berquerant/cron2date/internal/mcp"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func connectInMemory(t *testing.T, opts cronmcp.Options) *mcp.ClientSession {
	t.Helper()
	ctx := context.Background()
	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	server := cronmcp.NewServer(opts)
	_, err := server.Connect(ctx, serverTransport, nil)
	require.NoError(t, err)

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "0.1.0"}, nil)
	session, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = session.Close()
	})
	return session
}

func decodeStructured[T any](t *testing.T, res *mcp.CallToolResult) T {
	t.Helper()
	raw, err := json.Marshal(res.StructuredContent)
	require.NoError(t, err)
	var out T
	require.NoError(t, json.Unmarshal(raw, &out))
	return out
}

func TestMCPServer(t *testing.T) {
	ctx := context.Background()
	fixedNow := func() time.Time {
		return time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	}
	session := connectInMemory(t, cronmcp.Options{
		Version:    "test",
		MaxResults: 5,
		Now:        fixedNow,
	})

	t.Run("list tools", func(t *testing.T) {
		res, err := session.ListTools(ctx, &mcp.ListToolsParams{})
		require.NoError(t, err)
		names := make([]string, 0, len(res.Tools))
		for _, tool := range res.Tools {
			names = append(names, tool.Name)
			require.NotNil(t, tool.Annotations)
			assert.True(t, tool.Annotations.ReadOnlyHint)
			assert.True(t, tool.Annotations.IdempotentHint)
		}
		assert.ElementsMatch(t, []string{"expand_cron", "validate_cron"}, names)
	})

	t.Run("expand_cron success", func(t *testing.T) {
		res, err := session.CallTool(ctx, &mcp.CallToolParams{
			Name: "expand_cron",
			Arguments: map[string]any{
				"expr":     "0 * * * *",
				"start":    "2026-06-01 10:00:00",
				"count":    3,
				"timezone": "UTC",
			},
		})
		require.NoError(t, err)
		assert.False(t, res.IsError)
		out := decodeStructured[cronmcp.ExpandCronOutput](t, res)
		assert.Equal(t, cronmcp.ExpandCronOutput{
			Times: []string{
				"2026-06-01 11:00:00",
				"2026-06-01 12:00:00",
				"2026-06-01 13:00:00",
			},
			Count:     3,
			Timezone:  "UTC",
			Truncated: false,
		}, out)
	})

	t.Run("expand_cron with inclusive and format preset", func(t *testing.T) {
		res, err := session.CallTool(ctx, &mcp.CallToolParams{
			Name: "expand_cron",
			Arguments: map[string]any{
				"expr":      "0 9 * * *",
				"start":     "2026-06-01T00:00:00Z",
				"count":     2,
				"format":    "rfc3339",
				"timezone":  "Asia/Tokyo",
				"inclusive": true,
			},
		})
		require.NoError(t, err)
		assert.False(t, res.IsError)
		out := decodeStructured[cronmcp.ExpandCronOutput](t, res)
		assert.Equal(t, []string{
			"2026-06-01T09:00:00+09:00",
			"2026-06-02T09:00:00+09:00",
		}, out.Times)
		assert.Equal(t, "Asia/Tokyo", out.Timezone)
	})

	t.Run("expand_cron with seconds and duration", func(t *testing.T) {
		res, err := session.CallTool(ctx, &mcp.CallToolParams{
			Name: "expand_cron",
			Arguments: map[string]any{
				"expr":     "*/15 * * * * *",
				"start":    "2026-06-01 10:00:00",
				"duration": "35s",
				"timezone": "UTC",
				"seconds":  true,
			},
		})
		require.NoError(t, err)
		assert.False(t, res.IsError)
		out := decodeStructured[cronmcp.ExpandCronOutput](t, res)
		assert.Equal(t, []string{"2026-06-01 10:00:15", "2026-06-01 10:00:30"}, out.Times)
		assert.Equal(t, 2, out.Count)
		assert.False(t, out.Truncated)
	})

	t.Run("expand_cron enforces max results cap", func(t *testing.T) {
		res, err := session.CallTool(ctx, &mcp.CallToolParams{
			Name: "expand_cron",
			Arguments: map[string]any{
				"expr":     "* * * * *",
				"duration": "1h",
				"timezone": "UTC",
			},
		})
		require.NoError(t, err)
		assert.False(t, res.IsError)
		out := decodeStructured[cronmcp.ExpandCronOutput](t, res)
		assert.Equal(t, 5, out.Count)
		assert.Len(t, out.Times, 5)
		assert.True(t, out.Truncated)
	})

	t.Run("expand_cron returns tool error on invalid inputs", func(t *testing.T) {
		for _, tc := range []struct {
			title string
			args  map[string]any
		}{
			{
				title: "empty expr",
				args:  map[string]any{"expr": "", "count": 1},
			},
			{
				title: "missing stop condition",
				args:  map[string]any{"expr": "0 * * * *"},
			},
			{
				title: "negative count",
				args:  map[string]any{"expr": "0 * * * *", "count": -1},
			},
			{
				title: "invalid duration",
				args:  map[string]any{"expr": "0 * * * *", "duration": "bad"},
			},
			{
				title: "negative duration",
				args:  map[string]any{"expr": "0 * * * *", "duration": "-1h"},
			},
			{
				title: "invalid timezone",
				args:  map[string]any{"expr": "0 * * * *", "count": 1, "timezone": "Invalid/Zone"},
			},
			{
				title: "invalid start",
				args:  map[string]any{"expr": "0 * * * *", "count": 1, "start": "bad"},
			},
			{
				title: "invalid cron expr",
				args:  map[string]any{"expr": "not-a-cron", "count": 1},
			},
		} {
			t.Run(tc.title, func(t *testing.T) {
				res, err := session.CallTool(ctx, &mcp.CallToolParams{
					Name:      "expand_cron",
					Arguments: tc.args,
				})
				require.NoError(t, err)
				assert.True(t, res.IsError)
			})
		}
	})

	t.Run("validate_cron valid and invalid expressions", func(t *testing.T) {
		res, err := session.CallTool(ctx, &mcp.CallToolParams{
			Name: "validate_cron",
			Arguments: map[string]any{
				"expr":     "0 12 * * *",
				"timezone": "Asia/Tokyo",
				"format":   time.RFC3339,
			},
		})
		require.NoError(t, err)
		assert.False(t, res.IsError)
		out := decodeStructured[cronmcp.ValidateCronOutput](t, res)
		assert.True(t, out.Valid)
		assert.Empty(t, out.Error)
		assert.Equal(t, "Asia/Tokyo", out.Timezone)
		assert.Equal(t, "2026-06-02T12:00:00+09:00", out.NextFire)

		invalidRes, err := session.CallTool(ctx, &mcp.CallToolParams{
			Name: "validate_cron",
			Arguments: map[string]any{
				"expr":     "invalid",
				"timezone": "UTC",
			},
		})
		require.NoError(t, err)
		assert.False(t, invalidRes.IsError)
		invalidOut := decodeStructured[cronmcp.ValidateCronOutput](t, invalidRes)
		assert.False(t, invalidOut.Valid)
		assert.NotEmpty(t, invalidOut.Error)
		assert.Empty(t, invalidOut.NextFire)

		unmatchableRes, err := session.CallTool(ctx, &mcp.CallToolParams{
			Name: "validate_cron",
			Arguments: map[string]any{
				"expr":     "0 0 31 2 *",
				"timezone": "UTC",
			},
		})
		require.NoError(t, err)
		assert.False(t, unmatchableRes.IsError)
		unmatchableOut := decodeStructured[cronmcp.ValidateCronOutput](t, unmatchableRes)
		assert.True(t, unmatchableOut.Valid)
		assert.Empty(t, unmatchableOut.NextFire)

		secRes, err := session.CallTool(ctx, &mcp.CallToolParams{
			Name: "validate_cron",
			Arguments: map[string]any{
				"expr":     "*/15 * * * * *",
				"start":    "2026-06-01 10:00:00",
				"timezone": "UTC",
				"seconds":  true,
			},
		})
		require.NoError(t, err)
		assert.False(t, secRes.IsError)
		secOut := decodeStructured[cronmcp.ValidateCronOutput](t, secRes)
		assert.True(t, secOut.Valid)
		assert.Equal(t, "2026-06-01 10:00:15", secOut.NextFire)

		cronTZRes, err := session.CallTool(ctx, &mcp.CallToolParams{
			Name: "validate_cron",
			Arguments: map[string]any{
				"expr":     "CRON_TZ=UTC 0 12 * * *",
				"timezone": "Asia/Tokyo",
				"format":   time.RFC3339,
			},
		})
		require.NoError(t, err)
		assert.False(t, cronTZRes.IsError)
		cronTZOut := decodeStructured[cronmcp.ValidateCronOutput](t, cronTZRes)
		assert.True(t, cronTZOut.Valid)
		assert.Equal(t, "Asia/Tokyo", cronTZOut.Timezone)
		assert.Equal(t, "2026-06-01T21:00:00+09:00", cronTZOut.NextFire)

		for _, badArgs := range []map[string]any{
			{"expr": "0 * * * *", "timezone": "Invalid/Zone"},
			{"expr": "0 * * * *", "start": "not-a-time"},
		} {
			errRes, err := session.CallTool(ctx, &mcp.CallToolParams{
				Name:      "validate_cron",
				Arguments: badArgs,
			})
			require.NoError(t, err)
			assert.True(t, errRes.IsError)
		}
	})
}
