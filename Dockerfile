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

# Install Cairo development dependencies
RUN apt-get update && apt-get install -y \
    libcairo2-dev \
    pkg-config \
    && rm -rf /var/lib/apt/lists/*

COPY . .
COPY --from=uibuilder /src/dist ./ui/dist
RUN go generate ./... && go build -tags cairo -ldflags "-s -w -X main.version=${VERSION}" -o rmfakecloud-docker ./cmd/rmfakecloud/

FROM debian:bookworm-slim
EXPOSE 3000

# Install Cairo runtime libraries
RUN apt-get update && apt-get install -y \
    ca-certificates \
    libcairo2 \
    && rm -rf /var/lib/apt/lists/*

ADD ./docker/rootfs.tar /
COPY --from=gobuilder /src/rmfakecloud-docker /
ENTRYPOINT ["/rmfakecloud-docker"]
