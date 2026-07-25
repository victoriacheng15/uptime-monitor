# =============================================================================
# LOCAL DEV ENVIRONMENT - Silverblue (Podman, no compose available)
#
# Build:
#   podman build -t uptime-monitor-dev .
#
# Run:
#   podman run -it --rm \
#     --name uptime-monitor-dev \
#     -p 8080:8080 \
#     -v "$(pwd):/workspace:Z" \
#     -e MONITOR_TARGETS="https://google.com,https://github.com" \
#     uptime-monitor-dev
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

# Build Air and goimports from source
RUN go install github.com/air-verse/air@latest && \
    go install golang.org/x/tools/cmd/goimports@latest

# Download Tailwind CSS CLI binary
RUN curl -sL https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-x64 \
    -o /usr/local/bin/tailwindcss && chmod +x /usr/local/bin/tailwindcss

# =============================================================================
# STAGE 2: dev
# The actual dev runtime. Only copies built binaries from the tools stage.
# Keeps make, bash, and git for Air's build commands and shell scripts.
# =============================================================================
FROM golang:1.26-alpine AS dev

# Install dev runtime requirements and compatibility packages
RUN apk add --no-cache make bash git gcompat libgcc libstdc++

# Copy tooling binaries from tools stage
COPY --from=tools /go/bin/air                /usr/local/bin/air
COPY --from=tools /go/bin/goimports          /usr/local/bin/goimports
COPY --from=tools /usr/local/bin/tailwindcss /usr/local/bin/tailwindcss

WORKDIR /workspace
COPY go.mod go.sum ./
RUN go mod download

EXPOSE 8080

# Air watches all .go files and restarts the dev server on any change.
CMD ["air"]
