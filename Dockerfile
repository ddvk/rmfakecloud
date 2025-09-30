ARG VERSION=0.0.0
FROM --platform=$BUILDPLATFORM node:lts AS uibuilder
ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
RUN corepack enable pnpm && corepack install -g pnpm@latest-9

WORKDIR /src
#COPY ui/package.json ui/pnpm-lock.yaml /src
#RUN pnpm fetch 

COPY ui .
RUN pnpm install && pnpm build

FROM golang:bookworm AS gobuilder
ARG VERSION
WORKDIR /src
COPY . .
COPY --from=uibuilder /src/dist ./ui/dist
#RUN apk add git
RUN go generate ./... && CGO_ENABLED=0 go build -ldflags "-s -w -X main.version=${VERSION}" -o rmfakecloud-docker ./cmd/rmfakecloud/

# Build Python + rmc + Inkscape stage for v6 support
FROM python:3.11-slim AS rmcbuilder
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        inkscape \
    && rm -rf /var/lib/apt/lists/*
RUN pip install --no-cache-dir rmc

FROM debian:bookworm-slim
EXPOSE 3000

# Install runtime dependencies for Python and Inkscape
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        ca-certificates \
        libpython3.11 \
        inkscape \
    && rm -rf /var/lib/apt/lists/*

# Copy Python from rmcbuilder
COPY --from=rmcbuilder /usr/local/lib/python3.11 /usr/local/lib/python3.11
COPY --from=rmcbuilder /usr/local/bin/python3.11 /usr/local/bin/python3.11
COPY --from=rmcbuilder /usr/local/bin/rmc /usr/local/bin/rmc

# Create symlinks for python
RUN ln -s /usr/local/bin/python3.11 /usr/local/bin/python3 && \
    ln -s /usr/local/bin/python3.11 /usr/local/bin/python

# Copy rmfakecloud binary
COPY --from=gobuilder /src/rmfakecloud-docker /rmfakecloud

# Set environment for v6 support
ENV RMC_PATH=/usr/local/bin/rmc
ENV INKSCAPE_PATH=/usr/bin/inkscape
ENV RMC_TIMEOUT=60

ENTRYPOINT ["/rmfakecloud"]
