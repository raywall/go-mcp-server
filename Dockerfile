FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY ./app .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o mcp-server server.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/mcp-server .
ENTRYPOINT ["./mcp-server"]