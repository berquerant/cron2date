package mcp

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/berquerant/cron2date/internal/cron2date"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// DefaultMaxResults caps the number of timestamps returned in a single expand_cron call
// to prevent memory exhaustion and oversized tool responses.
const DefaultMaxResults = 1000

// Options configures the cron2date MCP server.
type Options struct {
	Version    string
	MaxResults int
	Now        func() time.Time
}

// ExpandCronInput defines the parameters for the expand_cron tool.
type ExpandCronInput struct {
	Expr     string `json:"expr" jsonschema:"Cron expression (5-field standard, descriptor, or 6-field with seconds=true)"`
	Start    string `json:"start,omitempty" jsonschema:"Start timestamp in format (default: current time)"`
	End      string `json:"end,omitempty" jsonschema:"End timestamp in format"`
	Count    *int   `json:"count,omitempty" jsonschema:"Maximum number of firing times to return (0 or greater)"`
	Duration string `json:"duration,omitempty" jsonschema:"Duration from start time (Go duration, e.g. 24h, 3h30m)"`
	Format   string `json:"format,omitempty" jsonschema:"Go time layout (default: 2006-01-02 15:04:05)"`
	Timezone string `json:"timezone,omitempty" jsonschema:"IANA timezone name (e.g. UTC, Asia/Tokyo; default: local)"`
	Seconds  bool   `json:"seconds,omitempty" jsonschema:"Parse 6-field cron expression with leading seconds field"`
}

// ExpandCronOutput defines the structured response for the expand_cron tool.
type ExpandCronOutput struct {
	Times     []string `json:"times" jsonschema:"Expanded firing timestamps"`
	Count     int      `json:"count" jsonschema:"Number of returned timestamps"`
	Timezone  string   `json:"timezone" jsonschema:"Effective timezone used"`
	Truncated bool     `json:"truncated" jsonschema:"True if output was capped at the server maximum limit"`
}

// ValidateCronInput defines the parameters for the validate_cron tool.
type ValidateCronInput struct {
	Expr     string `json:"expr" jsonschema:"Cron expression to validate"`
	Start    string `json:"start,omitempty" jsonschema:"Reference timestamp for next firing preview (default: now)"`
	Format   string `json:"format,omitempty" jsonschema:"Go time layout (default: 2006-01-02 15:04:05)"`
	Timezone string `json:"timezone,omitempty" jsonschema:"IANA timezone name (e.g. UTC, Asia/Tokyo; default: local)"`
	Seconds  bool   `json:"seconds,omitempty" jsonschema:"Validate as a 6-field cron expression with leading seconds"`
}

// ValidateCronOutput defines the structured response for the validate_cron tool.
type ValidateCronOutput struct {
	Valid    bool   `json:"valid" jsonschema:"Whether the cron expression is valid"`
	Error    string `json:"error,omitempty" jsonschema:"Parse error message when valid is false"`
	NextFire string `json:"next_fire,omitempty" jsonschema:"Next firing time after start, empty if never fires"`
	Timezone string `json:"timezone" jsonschema:"Effective timezone used"`
}

// NewServer creates a configured cron2date MCP server exposing expand_cron and validate_cron.
func NewServer(opts Options) *mcp.Server {
	if opts.Version == "" {
		opts.Version = cron2date.DefaultVersion
	}
	if opts.MaxResults <= 0 {
		opts.MaxResults = DefaultMaxResults
	}
	if opts.Now == nil {
		opts.Now = time.Now
	}

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "cron2date",
		Version: opts.Version,
	}, nil)

	h := &handlers{
		maxResults: opts.MaxResults,
		now:        opts.Now,
	}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "expand_cron",
		Description: "Expand a cron expression into concrete firing timestamps given count, duration, or end time.",
		Annotations: &mcp.ToolAnnotations{
			Title:          "Expand Cron Expression",
			ReadOnlyHint:   true,
			IdempotentHint: true,
			OpenWorldHint:  new(false),
		},
	}, h.expandCron)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "validate_cron",
		Description: "Validate a cron expression and preview its next firing timestamp.",
		Annotations: &mcp.ToolAnnotations{
			Title:          "Validate Cron Expression",
			ReadOnlyHint:   true,
			IdempotentHint: true,
			OpenWorldHint:  new(false),
		},
	}, h.validateCron)

	return server
}

// RunStdio runs the cron2date MCP server over standard input/output.
func RunStdio(ctx context.Context, opts Options) error {
	return NewServer(opts).Run(ctx, &mcp.StdioTransport{})
}

type handlers struct {
	maxResults int
	now        func() time.Time
}

func toExpandParams(in ExpandCronInput) (cron2date.ExpandParams, error) {
	params := cron2date.ExpandParams{
		Expr:     in.Expr,
		Start:    in.Start,
		End:      in.End,
		Count:    -1,
		Duration: -1,
		Format:   in.Format,
		Timezone: in.Timezone,
		Seconds:  in.Seconds,
	}

	if in.Count != nil {
		if *in.Count < 0 {
			return params, errors.New("count must be 0 or greater")
		}
		params.Count = *in.Count
	}

	if in.Duration != "" {
		d, err := time.ParseDuration(in.Duration)
		if err != nil {
			return params, fmt.Errorf("parse duration %s: %w", in.Duration, err)
		}
		if d < 0 {
			return params, errors.New("duration must be 0 or greater")
		}
		params.Duration = d
	}

	return params, nil
}

func (h *handlers) expandCron(
	ctx context.Context,
	_ *mcp.CallToolRequest,
	in ExpandCronInput,
) (*mcp.CallToolResult, ExpandCronOutput, error) {
	params, err := toExpandParams(in)
	if err != nil {
		return nil, ExpandCronOutput{}, err
	}

	prepared, err := cron2date.PrepareExpand(params, h.now)
	if err != nil {
		return nil, ExpandCronOutput{}, err
	}

	times := make([]string, 0)
	truncated := false
	for x := range prepared.Seq {
		if err := ctx.Err(); err != nil {
			return nil, ExpandCronOutput{}, err
		}
		if len(times) >= h.maxResults {
			truncated = true
			break
		}
		times = append(times, x.In(prepared.Location).Format(prepared.Format))
	}

	return nil, ExpandCronOutput{
		Times:     times,
		Count:     len(times),
		Timezone:  prepared.Location.String(),
		Truncated: truncated,
	}, nil
}

func (h *handlers) validateCron(
	_ context.Context,
	_ *mcp.CallToolRequest,
	in ValidateCronInput,
) (*mcp.CallToolResult, ValidateCronOutput, error) {
	loc, err := cron2date.ResolveLocation(in.Timezone)
	if err != nil {
		return nil, ValidateCronOutput{}, err
	}
	layout := cron2date.ResolveFormat(in.Format)

	start, err := cron2date.ParseTimeOrDefault(in.Start, layout, loc, h.now().In(loc))
	if err != nil {
		return nil, ValidateCronOutput{}, err
	}

	nexter, err := cron2date.ParseNexter(in.Expr, in.Seconds)
	if err != nil {
		return nil, ValidateCronOutput{
			Valid:    false,
			Error:    err.Error(),
			Timezone: loc.String(),
		}, nil
	}

	next := nexter.Next(start)
	nextFire := ""
	if !next.IsZero() {
		nextFire = next.In(loc).Format(layout)
	}

	return nil, ValidateCronOutput{
		Valid:    true,
		NextFire: nextFire,
		Timezone: loc.String(),
	}, nil
}
