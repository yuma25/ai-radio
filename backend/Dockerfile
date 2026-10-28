# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy dependency definitions
COPY go.mod ./
# COPY go.sum ./ (if exists)
RUN go mod download || true

# Copy source code
COPY . .

# Build executable
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/08_server/main.go

# Runtime stage
FROM alpine:3.21

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /app/server /app/server

EXPOSE 8080

CMD ["/app/server"]
