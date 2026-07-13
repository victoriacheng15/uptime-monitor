# =============================================================================
# LOCAL DEV ENVIRONMENT - Silverblue (Podman, no compose available)
#
# Build:
#   podman build -t uptime-monitor-dev .
#
# Run:
# podman run -it --rm \
#   -p 8080:8080 \
#   -v "$(pwd):/app:Z" \
#   -e MONITOR_TARGETS="https://google.com,https://github.com" \
#   uptime-monitor-dev
#
# Note: The :Z flag on the volume mount is required on SELinux/Silverblue
#       systems to relabel the volume so the container can read/write it.
#
# Open http://localhost:8080 in your browser.
# =============================================================================

# =============================================================================
# STAGE 1: tools
# Installs curl and git, downloads Air and the Tailwind CSS CLI binary.
# Nothing from this stage carries over to the final dev image except binaries.
# =============================================================================
FROM golang:1.26-alpine AS tools

RUN apk add --no-cache curl git

# Build Air from source into /go/bin/air
RUN go install github.com/air-verse/air@latest

# Download Tailwind CSS CLI binary
RUN curl -sL https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-x64 \
    -o /usr/local/bin/tailwindcss && chmod +x /usr/local/bin/tailwindcss

# =============================================================================
# STAGE 2: dev
# The actual dev runtime. Only copies built binaries from the tools stage.
# Keeps make, bash, and git for Air's build commands and shell scripts.
# =============================================================================
FROM golang:1.26-alpine AS dev

RUN apk add --no-cache make bash git

# Copy tooling binaries from tools stage
COPY --from=tools /go/bin/air               /usr/local/bin/air
COPY --from=tools /usr/local/bin/tailwindcss /usr/local/bin/tailwindcss

WORKDIR /app

EXPOSE 8080

# Air watches all .go files and restarts the dev server on any change.
# cmd/server/main.go handles:
#   - Backend:  Go API router (health, check, latest, history)
#   - Backend:  S3 persistence overridden with local JSON file storage in dist/
#   - Frontend: Go SSG (web.Generator) + Tailwind CSS compiled on startup
#   - Frontend: internal/web/templates polled every 500ms, rebuilt on any edit
#   - Frontend: SSE live-reload script injected into HTML pages
CMD ["air"]
