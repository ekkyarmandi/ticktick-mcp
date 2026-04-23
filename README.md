# TickTick MCP

A Go-based Model Context Protocol (MCP) server for TickTick. It exposes TickTick project and task operations over stdio using the official Go MCP SDK.

## Overview

This server lets MCP clients interact with TickTick through tools for:

- Retrieving projects and project data
- Retrieving task details and tasks due today
- Creating projects and tasks
- Updating tasks
- Completing and deleting tasks

## Requirements

- Go 1.25+
- TickTick account
- TickTick OpenAPI access token

## Installation

1. Clone this repository.

   ```bash
   git clone https://github.com/ekkyarmandi/ticktick-mcp.git
   cd ticktick-mcp
   ```

2. Download Go dependencies.

   ```bash
   go mod tidy
   ```

## TickTick Authentication

This server uses TickTick's OpenAPI bearer token. Register an app in the TickTick developer portal, complete the OAuth flow there, and provide the resulting access token to this server.

## Configuration

Set the following environment variable before starting the server:

```bash
export TICKTICK_API_KEY=your_access_token_here
```

Optional:

```bash
export TICKTICK_API_BASE=https://api.ticktick.com/open/v1
```

## Usage

Run the MCP server over stdio:

```bash
go run .
```

If your environment restricts writes to Go's default cache directories, run it with cache paths under `/tmp`:

```bash
GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod-cache go run .
```

The server exposes these tools:

- `get_projects`
- `project_details`
- `get_today_tasks`
- `get_task_details`
- `create_project`
- `create_task`
- `update_task`
- `complete_task`
- `delete_task`

## Using With MCP Clients

Register the built binary or `go run .` command as a stdio MCP server in your client. Example use cases:

- "Show me all my TickTick projects"
- "Create a project named Home Renovation"
- "List tasks due today in Asia/Jakarta"
- "Mark my Pay bills task as complete"

### Claude Code Project Config

This repository includes a project-scoped [.mcp.json](/Users/ekkyarmandi/PARA/01_PROJECTS/Personal/mcp/ticktick-mcp/.mcp.json). A working configuration looks like this:

```json
{
  "mcpServers": {
    "ticktick-mcp": {
      "type": "stdio",
      "command": "go",
      "args": ["run", "."],
      "cwd": ".",
      "env": {
        "TICKTICK_API_KEY": "YOUR_TICKTICK_API_KEY",
        "GOCACHE": "/tmp/go-build",
        "GOMODCACHE": "/tmp/go-mod-cache"
      }
    }
  }
}
```

`GOCACHE` and `GOMODCACHE` are included because some sandboxed environments cannot write to Go's default cache locations, which causes `claude mcp get` or `claude mcp list` health checks to fail before the MCP handshake.

## Development

The Go entrypoint is `main.go`, and TickTick client plus tool handlers live in `ticktick.go`.

## License

MIT
