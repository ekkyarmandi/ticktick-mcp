FROM golang:1.25-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/ticktick-mcp .

FROM alpine:3.22

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=build /out/ticktick-mcp /app/ticktick-mcp

ENV MCP_TRANSPORT=http
ENV PORT=8080
ENV MCP_HTTP_PATH=/mcp

EXPOSE 8080

ENTRYPOINT ["/app/ticktick-mcp"]
