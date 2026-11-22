# cron2date

CLI tool to expand cron expressions into concrete firing timestamps.

## Install

```sh
go install github.com/berquerant/cron2date/cmd/cron2date@latest
```

## Usage

```
❯ cron2date -h
cron2date - expand cron expression

Usage:

  cron2date [flags] CRONEXPR

Flags:
  -c int
    	firing times to display, negative value is ignored (default -1)
  -d duration
    	set the end time to the start time + this duration, negative value is ignored (default -1ns)
  -e string
    	end time
  -f string
    	time layout (default "2006-01-02 15:04:05")
  -mcp
    	run as an MCP server over stdio
  -s string
    	start time (default current time)
  -seconds
    	parse 6-field cron expression with leading seconds
  -tz string
    	timezone name (e.g. UTC, Asia/Tokyo; default local)
  -version
    	print version and exit
```

## Examples

```sh
# Next 5 firing times from now
cron2date -c 5 "0 * * * *"

# Firing times within a duration from a given start time
cron2date -s "2026-06-01 10:00:00" -d 3h30m "0 * * * *"

# Firing times in a specific timezone using RFC3339 output
cron2date -tz Asia/Tokyo -f "2006-01-02T15:04:05Z07:00" -c 3 "0 9 * * *"

# 6-field cron expression with seconds
cron2date -seconds -s "2026-06-01 10:00:00" -c 3 "*/15 * * * * *"

# Descriptors and CRON_TZ prefix (supported via robfig/cron/v3)
cron2date -c 3 "@daily"
cron2date -c 3 "CRON_TZ=UTC 0 0 * * *"
```

## MCP Server

Run `cron2date -mcp` to start a stdio MCP server exposing two read-only tools:

- `expand_cron`: expand a cron expression into concrete firing timestamps (capped at 1000 results per call).
- `validate_cron`: validate a cron expression and preview its next firing timestamp.

Example MCP client configuration:

```json
{
  "mcpServers": {
    "cron2date": {
      "command": "cron2date",
      "args": ["-mcp"]
    }
  }
}
```

## License

[MIT](LICENSE)
