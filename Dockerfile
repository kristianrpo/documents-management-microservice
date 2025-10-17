# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install tools needed for fetching Go modules (some modules require git)
RUN apk --no-cache add git

# Copy go.mod first to leverage build cache for dependencies
COPY go.mod ./
RUN go mod download

# Copy the full source
COPY . .

# Ensure go.sum is generated and all deps are resolved
RUN go mod tidy

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o server cmd/server/main.go

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/server .

# Expose port
EXPOSE 8080

# Run the application
CMD ["./server"]
