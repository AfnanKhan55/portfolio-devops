# ---- Stage 1: build ---------------------------------------------------
# Uses the full Go toolchain to compile a static binary. This image is
# ~800MB but it NEVER ships — it's discarded after the build finishes.
FROM golang:1.22-alpine AS builder

WORKDIR /src

# Copy dependency manifests first so Docker can cache this layer and skip
# re-downloading modules when only application code changes.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO_ENABLED=0 -> fully static binary (no libc dependency), which is what
# lets us run it on a "distroless" image that has no shared libraries at all.
ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X github.com/AfnanKhan55/portfolio-devops/internal/handlers.Version=${VERSION}" \
    -o /out/server ./cmd/server

# ---- Stage 2: runtime ---------------------------------------------------
# distroless/static has no shell, no package manager, no OS utilities —
# just enough to run a static binary. Much smaller attack surface than
# even alpine, and there's nothing for an attacker to "get a shell" with.
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY --from=builder /out/server ./server
COPY --from=builder /src/web ./web

# distroless "nonroot" images already run as an unprivileged user (uid 65532),
# so the container never runs as root by default.
USER nonroot:nonroot

EXPOSE 8080
ENTRYPOINT ["./server"]
