# ============================================================================
# Build stage: Generate templ files, build Tailwind CSS, compile Go binary
# ============================================================================
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache nodejs npm git

WORKDIR /app

# Install Go dependencies
COPY go.mod go.sum ./
RUN go mod download

# Install templ
RUN go install github.com/a-h/templ/cmd/templ@latest

# Copy source code
COPY . .

# Generate templ files
RUN templ generate

# Build Tailwind CSS
RUN npm ci && npm run css:build

# Build Go binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/server

# ============================================================================
# Runtime stage: Minimal image with just the binary and static assets
# ============================================================================
FROM alpine:3.20

RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy binary and static files
COPY --from=builder /app/server .
COPY --from=builder /app/web/static ./web/static

EXPOSE 8080

CMD ["./server"]
