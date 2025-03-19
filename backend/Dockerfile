FROM golang:1.18-alpine as builder

# Install dependencies
RUN apk --no-cache add ca-certificates git make gcc libc-dev

WORKDIR /app

# Copy and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o /go/bin/codehawk-api ./cmd/server

# Create a minimal production image
FROM alpine:3.15

# Install required packages
RUN apk --no-cache add ca-certificates tzdata

# Copy binary from builder
COPY --from=builder /go/bin/codehawk-api /app/codehawk-api

# Create a non-root user to run the application
RUN adduser -D -g '' appuser
USER appuser

# Set working directory
WORKDIR /app

# Set environment variables
ENV GIN_MODE=release
ENV PORT=8080

# Expose port
EXPOSE 8080

# Run the application
CMD ["./codehawk-api"]