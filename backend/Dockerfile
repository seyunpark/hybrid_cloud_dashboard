FROM golang:1.25-alpine AS builder

WORKDIR /app

# CGO 활성화 (go-sqlite3 빌드에 필요)
RUN apk --no-cache add gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -o server ./cmd/server

FROM alpine:3.21

RUN apk --no-cache add ca-certificates curl

WORKDIR /app

COPY --from=builder /app/server .

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=10s --retries=3 --start-period=40s \
  CMD curl -f http://localhost:8080/health || exit 1

ENTRYPOINT ["./server"]
