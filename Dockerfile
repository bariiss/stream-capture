# Build stage
ARG GO_VERSION=1.25
ARG VCS_REF
ARG BUILD_DATE
ARG VERSION

FROM golang:${GO_VERSION}-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum* ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
ARG TARGETARCH
ARG TARGETPLATFORM
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} \
    go build -a -installsuffix cgo \
    -ldflags "-w -s" \
    -o stream-capture ./cmd/stream-capture

# Final stage
FROM python:3.12-slim

ARG VCS_REF
ARG BUILD_DATE
ARG VERSION

LABEL org.opencontainers.image.title="stream-capture" \
      org.opencontainers.image.description="HLS stream capture tool" \
      org.opencontainers.image.vendor="bariiss" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.revision="${VCS_REF}" \
      org.opencontainers.image.created="${BUILD_DATE}" \
      org.opencontainers.image.source="https://github.com/bariiss/stream-capture"

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    ca-certificates \
    tzdata \
    ffmpeg && \
    pip3 install --no-cache-dir openai-whisper && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/stream-capture .

# Default command
ENTRYPOINT ["./stream-capture"]

