# =============================================================
# Stage 1: Builder
# Full Go image used only to compile the application binary.
# Nothing from this stage ends up in the final image except the binary.
# =============================================================
# golang:alpine always pulls the latest stable Go release on Docker Hub,
# ensuring it satisfies whatever go.mod minimum version is declared.
FROM golang:alpine AS builder

# git and ca-certificates required by some Go modules during download.
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copy dependency manifests first — Docker layer caching means this
# layer only re-runs when go.mod or go.sum change, not on every code change.
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code.
COPY . .

# Issue #4 fix: build the package directory, not a single .go file.
# CGO_ENABLED=0  — statically linked binary, no external lib deps.
# GOOS=linux     — target Linux (even when building on Mac/Windows).
# -o /app/server — output binary path.
# ./cmd/server   — the package to build (directory, not a file).
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/server ./cmd/server

# =============================================================
# Stage 2: Runner
# Issue #22 fix: updated Alpine from 3.18 (EOL) to 3.20.
# Only the compiled binary is copied from the builder stage.
# Final image is ~15 MB — no Go toolchain, no source code.
# =============================================================
FROM alpine:3.20

# ca-certificates for HTTPS connections; tzdata for timezone support.
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy ONLY the compiled binary from the builder stage.
COPY --from=builder /app/server .

EXPOSE 8080

CMD ["./server"]
