# TickTick MCP

A Go-based Model Context Protocol (MCP) server for TickTick. It supports local stdio mode for desktop MCP clients and Streamable HTTP mode for cloud deployment.

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

For a more reliable launch from Claude or from a different working directory, use a small wrapper script such as `start.sh`:

```bash
#!/bin/zsh
cd /absolute/path/to/ticktick-mcp
export GOCACHE=/tmp/go-build
export GOMODCACHE=/tmp/go-mod-cache
exec /opt/homebrew/bin/go run .
```

Run the MCP server over HTTP for remote/cloud access:

```bash
MCP_TRANSPORT=http PORT=8080 go run .
```

You can also customize the endpoint path and enable an optional shared token:

```bash
MCP_TRANSPORT=http \
PORT=8080 \
MCP_HTTP_PATH=/mcp \
MCP_SERVER_TOKEN=replace-me \
go run .
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

### Claude Code User Config

For a user-scoped or cross-project Claude MCP config, prefer a wrapper script over `args: ["run", "/absolute/path/to/ticktick-mcp"]`. The wrapper is more reliable because it starts `go run .` from the module root and sets writable cache directories first.

Example:

```json
{
  "mcpServers": {
    "ticktick-mcp": {
      "type": "stdio",
      "command": "/absolute/path/to/ticktick-mcp/start.sh",
      "args": [],
      "env": {
        "TICKTICK_API_KEY": "YOUR_TICKTICK_API_KEY"
      }
    }
  }
}
```

Using `go run /absolute/path/to/ticktick-mcp` directly may fail because it is not equivalent to starting `go run .` from inside the module directory.

### Remote HTTP MCP

For cloud deployment, run the server in HTTP mode and expose the MCP endpoint over HTTPS. The default endpoint path is `/mcp`, and the health endpoint is `/healthz`.

Example local test:

```bash
MCP_TRANSPORT=http PORT=8080 go run .
```

Then your remote MCP endpoint is:

```text
https://your-domain.example/mcp
```

If you set `MCP_SERVER_TOKEN`, clients can authenticate with any of these headers:

```text
Authorization: Bearer YOUR_TOKEN
X-API-Key: YOUR_TOKEN
Api-Key: YOUR_TOKEN
```

## Docker

Build the container image:

```bash
docker build -t ticktick-mcp .
```

Run it as a cloud-ready HTTP server:

```bash
docker run --rm -p 8080:8080 \
  -e TICKTICK_API_KEY=your_access_token_here \
  -e MCP_TRANSPORT=http \
  -e PORT=8080 \
  -e MCP_HTTP_PATH=/mcp \
  -e MCP_SERVER_TOKEN=replace-me \
  ticktick-mcp
```

Then verify:

```bash
curl http://localhost:8080/healthz
```

For public deployment, put the container behind HTTPS and keep `TICKTICK_API_KEY` and `MCP_SERVER_TOKEN` as server-side secrets.

## Development

The Go entrypoint and transport setup live in `main.go`, and TickTick client plus tool handlers live in `ticktick.go`.

## License

MIT
