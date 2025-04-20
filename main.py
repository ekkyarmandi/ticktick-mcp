# main.py
from mcp.server.fastmcp import FastMCP
from mcp.server.sse import SseServerTransport
from starlette.applications import Starlette
from starlette.routing import Mount, Route

# Import all tools
from tools import *

# Create MCP server
mcp = FastMCP("ticktick_mcp")

mcp.add_tool(get_projects)
mcp.add_tool(project_details)
mcp.add_tool(get_task_details)
mcp.add_tool(create_project)
mcp.add_tool(create_task)
mcp.add_tool(update_task)
mcp.add_tool(complete_task)
mcp.add_tool(delete_task)

# Create SSE transport
tp = SseServerTransport("/messages/")


# Define handler functions for each server
async def handle_sse(request):
    async with tp.connect_sse(request.scope, request.receive, request._send) as streams:
        await mcp._mcp_server.run(
            streams[0],
            streams[1],
            mcp._mcp_server.create_initialization_options(),
        )


# Create a single Starlette app with different routes
routes = [
    Route("/sse", endpoint=handle_sse),
    Mount("/messages/", app=tp.handle_post_message),
]

# Single Starlette app
app = Starlette(routes=routes)

# Run the server
if __name__ == "__main__":
    import uvicorn

    uvicorn.run(app, host="0.0.0.0", port=8000)
