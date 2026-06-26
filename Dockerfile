# =============================================================
# Stage 1: Builder
# We use the full Go image to compile the application.
# This stage produces a single compiled binary.
# =============================================================
FROM golang:1.26.4-alpine AS builder

# Install git — required by some Go modules during download.
# ca-certificates — required for HTTPS connections during go mod download.
RUN apk add --no-cache git ca-certificates

# Set the working directory inside the container.
WORKDIR /app

# Copy go.mod and go.sum FIRST — before copying source code.
# Why? Docker layer caching. If only source code changes (not dependencies),
# Docker reuses the cached "go mod download" layer. This makes rebuilds fast.
# If we copied everything first, any code change would re-download all dependencies.
COPY go.mod go.sum ./

# Download all dependencies.
# This is cached as a Docker layer — only re-runs when go.mod/go.sum change.
RUN go mod download

# Now copy the rest of the source code.
COPY . .

# Build the application binary.
# CGO_ENABLED=0 — disable CGo, creating a statically linked binary.
#   This means the binary has no external library dependencies.
# GOOS=linux — target OS is Linux (required even if building on Mac/Windows).
# -o /app/server — output the binary to /app/server
# ./cmd/main.go — the entry point to compile
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/server ./cmd/main.go

# =============================================================
# Stage 2: Runner
# We use a minimal Alpine image — it's only ~5MB.
# We copy ONLY the compiled binary from the builder stage.
# The final image has no Go toolchain, no source code, no build tools.
# Result: a tiny, secure production image (~15MB total).
# Without multi-stage: the image would be ~800MB+ with the full Go toolchain.
# =============================================================
FROM alpine:3.18

# Install ca-certificates for HTTPS connections (e.g., to external APIs).
# tzdata for timezone support.
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy ONLY the compiled binary from the builder stage.
# --from=builder references the first stage by its alias name.
COPY --from=builder /app/server .

# Expose port 8080 — this documents which port the container listens on.
# It does NOT actually publish the port — that's done with -p in docker run.
EXPOSE 8080

# CMD is the command that runs when the container starts.
# We run the compiled binary directly — no interpreter needed.
# This is Go's advantage: a single binary with zero runtime dependencies.
CMD ["./server"]