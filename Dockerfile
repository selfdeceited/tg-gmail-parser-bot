# Stage 1: Build
# golang:1.23-bookworm used as 1.26 is not yet published on Docker Hub.
# GOTOOLCHAIN=auto lets the Go toolchain manager download and use the version
# required by go.mod (1.26.1) automatically.
FROM golang:1.23-bookworm AS builder

ENV GOTOOLCHAIN=auto

WORKDIR /src

# Download dependencies first (layer-cached separately from source)
COPY go.mod go.sum ./
RUN go mod download

# Copy full source and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -o /app/bot ./cmd/bot

# Stage 2: Minimal runtime image
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app
COPY --from=builder /app/bot ./bot

# Run as non-root (distroless nonroot = uid 65532)
USER nonroot:nonroot

ENTRYPOINT ["/app/bot"]
