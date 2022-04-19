ARG VERSION=0.0.0
FROM --platform=$BUILDPLATFORM node:lts-alpine as uibuilder
WORKDIR /src
COPY ui .
RUN yarn && yarn build 

FROM golang:1-alpine as gobuilder
ARG VERSION
WORKDIR /src
COPY . .
COPY --from=uibuilder /src/build ./ui/build
RUN apk add git
RUN go generate ./... && CGO_ENABLED=0 go build -ldflags "-s -w -X main.version=${VERSION}" -o rmfakecloud-docker ./cmd/rmfakecloud/

FROM scratch
EXPOSE 3000
COPY --from=gobuilder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=gobuilder /src/rmfakecloud-docker /
ENTRYPOINT ["/rmfakecloud-docker"]
