docker run -d --name ticktick-mcp -p 8080:8080 \
  -e TICKTICK_API_KEY=your_ticktick_access_token \
  -e MCP_TRANSPORT=http \
  -e PORT=8080 \
  -e MCP_HTTP_PATH=/mcp \
  ticktick-mcp
