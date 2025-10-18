FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build arguments for version info
ARG VERSION
ARG GIT_COMMIT
ARG BUILD_ENV

# Set build-time environment variables
ENV VERSION=${VERSION}
ENV GIT_COMMIT=${GIT_COMMIT}
ENV BUILD_ENV=${BUILD_ENV}

# Build the application and matching worker
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o matchingworker ./cmd/matchingworker

# Use distroless image for smaller size and better security
FROM gcr.io/distroless/static-debian12:latest

# Accept build arguments for runtime
ARG VERSION
ARG GIT_COMMIT
ARG BUILD_ENV

# Set environment variables for runtime
ENV VERSION=${VERSION}
ENV GIT_COMMIT=${GIT_COMMIT}
ENV BUILD_ENV=${BUILD_ENV}

# Copy the binaries from builder stage
COPY --from=builder /app/main /main
COPY --from=builder /app/matchingworker /app/matchingworker

# Copy migration files
COPY --from=builder /app/migrations /migrations

# Expose port
EXPOSE 8080

# Run the binary
ENTRYPOINT ["/main"]
