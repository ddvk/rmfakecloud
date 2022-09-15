ARG VERSION=0.0.0
FROM --platform=$BUILDPLATFORM node:lts as uibuilder
WORKDIR /src
COPY new-ui .
RUN yarn && yarn build 

FROM golang:1-alpine as gobuilder
ARG VERSION
WORKDIR /src
RUN apk add git
COPY . .
COPY --from=uibuilder /src/dist ./new-ui/dist
RUN go generate ./... && CGO_ENABLED=0 go build -ldflags "-s -w -X main.version=${VERSION}" -o rmfakecloud-docker ./cmd/rmfakecloud/

FROM alpine
EXPOSE 3000
COPY --from=gobuilder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=gobuilder /src/rmfakecloud-docker /
CMD ["/rmfakecloud-docker"]
