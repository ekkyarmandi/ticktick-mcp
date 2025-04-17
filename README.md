# TickTick MCP

A Model Context Protocol (MCP) server that provides tools for integrating TickTick task management tools. Using Python and the MCP SDK.

## Overview

This repository contains a Model Context Protocol (MCP) server implementation for TickTick. It provides a standardized way for AI assistants and applications to interact with TickTick's task management functionality, allowing operations like:

- Retrieving projects and tasks
- Creating new projects and tasks
- Updating task details
- Completing and deleting tasks

With this MCP, AI systems can act as task masters to help manage your to-do lists and tasks in TickTick with natural language.

## Requirements

- Python 3.8+
- TickTick account
- TickTick API key (via OAuth) # COMMENT: I will add a tool to generate an API key from the TickTick developer portal

## Installation

1. Clone this repository

   ```bash
   git clone https://github.com/ekkyarmandi/ticktick-mcp.git
   cd ticktick-mcp
   ```

2. Install dependencies

   ```bash
   pip install -r requirements.txt
   ```

## Obtaining a TickTick API Key

This MCP uses TickTick's OpenAPI scheme, which requires registering an app through TickTick's developer portal:

1. Go to the [TickTick Developer Documentation](https://developer.ticktick.com/docs)
2. Click on `Manage Apps` in the top right corner and login with your TickTick credentials
3. Register a new app by clicking the `+App Name` button
4. Enter a name for your app (only required field)
5. Once created, you'll be able to see your `Client ID` and `Client Secret`
6. For the `OAuth Redirect URL`, enter a URL where you'll be redirected after authorization (e.g., `http://127.0.0.1:8080`)

### Authorizing Your App

After registering your app, use the [ticktick-py](https://github.com/lazeroffmichael/ticktick-py) library to get your access token:

```python
from ticktick.oauth2 import OAuth2

# Replace with your details from the developer portal
client_id = "YOUR_CLIENT_ID"
client_secret = "YOUR_CLIENT_SECRET"
redirect_uri = "YOUR_REDIRECT_URI"  # e.g., http://127.0.0.1:8080

auth_client = OAuth2(client_id=client_id,
                    client_secret=client_secret,
                    redirect_uri=redirect_uri)

# This will open a web browser for authorization
# Follow the instructions in the terminal to authorize
auth_client.get_access_token()
```

After authorizing, the access token will be saved to a `.token-oauth` file by default. You can extract the token from this file or use:

```python
print(auth_client.token_info["access_token"])
```

## Configuration

1. Create a `.env` file in the root directory with your TickTick API key:
   ```
   TICKTICK_API_KEY=your_access_token_here
   ```

## Usage

Run the MCP server:

```bash
python main.py
```

This will start the MCP server on port 8000. You can now connect to it using any MCP client.

### Available Tools

The server provides the following tools:

- `get_projects`: Get a list of all projects
- `project_details`: Get details of a specific project
- `get_task_details`: Get details of a specific task
- `create_project`: Create a new project
- `create_task`: Create a new task in a project
- `update_task`: Update an existing task
- `complete_task`: Mark a task as complete
- `delete_task`: Delete a task

### Example Interactions

Once your MCP server is running, AI systems can help manage your tasks with natural language commands like:

- "Show me all my projects"
- "Create a new project called 'Home Renovation'"
- "Add a task to buy groceries tomorrow"
- "Mark my 'Pay bills' task as complete"
- "What tasks do I have due this week?"
- "Delete the task about the canceled meeting"

## Using with MCP Clients

This server can be used with any MCP-compatible client, such as:

- Claude Desktop
- Cursor IDE
- Custom AI applications using MCP SDKs

## Development

To extend or modify this MCP server:

1. Add new tools in `tools.py`
2. Register them in `main.py` using `mcp.add_tool()`

## License

MIT

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
