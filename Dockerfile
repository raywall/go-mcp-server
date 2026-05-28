FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Compilação estática para garantir execução leve
RUN CGO_ENABLED=0 GOOS=linux go build -o mcp-server main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/mcp-server .
ENTRYPOINT ["./mcp-server"]