# STAGE 1: Build the binary
FROM golang:1.25-alpine AS builder

# Install git and ca-certificates (needed for some Go modules)
RUN apk add --no-cache git ca-certificates

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum first to leverage Docker cache
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application
# CGO_ENABLED=0 creates a static binary that runs anywhere
# -ldflags="-w -s" strips debug info to reduce file size
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /grpc-server ./cmd/server/main.go

# STAGE 2: Final Runtime Image
FROM alpine:latest

# Add a non-root user for security (Senior best practice)
RUN adduser -D appuser
USER appuser

WORKDIR /home/appuser

# Copy only the binary from the builder stage
COPY --from=builder /grpc-server .

# Expose the gRPC port
EXPOSE 50051

# Run the binary
CMD ["./grpc-server"]