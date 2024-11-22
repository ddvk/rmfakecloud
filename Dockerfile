ARG VERSION=0.0.0
FROM --platform=$BUILDPLATFORM node:lts-slim AS uibuilder
ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
RUN corepack enable

WORKDIR /src
COPY pnpm-lock.yaml /src
RUN pnpm fetch --prod

COPY ui .
RUN pnpm i && pnpm build

FROM golang:1-alpine AS gobuilder
ARG VERSION
WORKDIR /src
COPY . .
COPY --from=uibuilder /src/dist ./ui/dist
RUN apk add git
RUN go generate ./... && CGO_ENABLED=0 go build -ldflags "-s -w -X main.version=${VERSION}" -o rmfakecloud-docker ./cmd/rmfakecloud/

FROM scratch
EXPOSE 3000
ADD ./docker/rootfs.tar /
COPY --from=gobuilder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=gobuilder /src/rmfakecloud-docker /
ENTRYPOINT ["/rmfakecloud-docker"]
