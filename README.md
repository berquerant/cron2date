# cron2date

CLI tool and stdio MCP server to expand cron expressions into concrete firing timestamps.

## Install

```sh
go install github.com/berquerant/cron2date/cmd/cron2date@latest
```

## Usage

```text
❯ cron2date -h
cron2date - expand cron expression

Usage:

  cron2date [flags] CRONEXPR
  cron2date -mcp [-mcp-max-results N]

Flags:
  -c int
    	firing times to display, negative value is ignored (default -1)
  -d duration
    	set the end time to the start time + this duration, negative value is ignored (default -1ns)
  -e string
    	end time
  -f string
    	time layout or preset (datetime, date, time, rfc3339, rfc3339nano) (default "2006-01-02 15:04:05")
  -inclusive
    	include start time if it matches CRONEXPR
  -mcp
    	run as an MCP server over stdio
  -mcp-max-results int
    	maximum results per MCP expand_cron call (requires -mcp) (default 1000)
  -s string
    	start time (default current time)
  -seconds
    	parse 6-field cron expression with leading seconds
  -tz string
    	timezone name (e.g. UTC, Asia/Tokyo; default local)
  -version
    	print version and exit
```

By default, firing times are strictly after `-s` (`(start, end]`). Pass `-inclusive` to include `-s` itself when it matches `CRONEXPR`.

## Examples

```sh
# Next 5 firing times from now
cron2date -c 5 "0 * * * *"

# Firing times within a duration from a given start time (inclusive)
cron2date -inclusive -s "2026-06-01 10:00:00" -d 3h "0 * * * *"

# Firing times in a specific timezone using the rfc3339 preset
cron2date -tz Asia/Tokyo -f rfc3339 -c 3 "0 9 * * *"

# 6-field cron expression with seconds
cron2date -seconds -s "2026-06-01 10:00:00" -c 3 "*/15 * * * * *"

# Descriptors and CRON_TZ prefix (output converted to -tz / local timezone)
cron2date -c 3 "@daily"
cron2date -tz Asia/Tokyo -c 3 "CRON_TZ=UTC 0 0 * * *"
```

## MCP Server

Run `cron2date -mcp` to start a stdio MCP server exposing two read-only tools:

- `expand_cron`: expand a cron expression into concrete firing timestamps (capped at `-mcp-max-results`, default 1000).
  - Parameters: `expr` (required), `start`, `end`, `count`, `duration`, `format`, `timezone`, `seconds`, `inclusive`
- `validate_cron`: validate a cron expression and preview its next firing timestamp.
  - Parameters: `expr` (required), `start`, `format`, `timezone`, `seconds`

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
