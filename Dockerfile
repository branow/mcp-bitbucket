FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /mcp-bitbucket ./cmd

FROM alpine:3.21

RUN apk add --no-cache ca-certificates && \
    adduser -D -h /home/appuser appuser

COPY --from=builder /mcp-bitbucket /mcp-bitbucket

USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD wget -qO- http://localhost:8080/health || exit 1

ENTRYPOINT ["/mcp-bitbucket"]
