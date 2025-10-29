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

# Install Cairo development libraries for native rmc-go
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        libcairo2-dev \
        pkg-config \
    && rm -rf /var/lib/apt/lists/*

COPY . .
COPY --from=uibuilder /src/dist ./ui/dist

# Build with Cairo support (native rmc-go)
RUN go generate ./... && \
    CGO_ENABLED=1 go build -tags cairo -ldflags "-s -w -X main.version=${VERSION}" -o rmfakecloud-docker ./cmd/rmfakecloud/

# Final runtime image - use Debian slim instead of Python
FROM debian:bookworm-slim
EXPOSE 3000

# Install runtime dependencies for Cairo
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        ca-certificates \
        libcairo2 \
    && rm -rf /var/lib/apt/lists/*

# Copy rmfakecloud binary
COPY --from=gobuilder /src/rmfakecloud-docker /rmfakecloud

# Set environment for native rmc-go (Cairo renderer)
ENV USE_NATIVE_RMC=true
ENV RMC_TIMEOUT=60

ENTRYPOINT ["/rmfakecloud"]
