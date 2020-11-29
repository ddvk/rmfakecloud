ARG VERSION=0.0.0
FROM node:alpine as uibuilder
WORKDIR /src
COPY ui .
RUN npm i && \
    npm run build

FROM golang:1-alpine as gobuilder
ARG VERSION
WORKDIR /src
COPY . .
COPY --from=uibuilder /src/build ./ui
RUN CGO_ENABLED=0 go build -ldflags "-s -w -X main.version=${VERSION}" -o rmfakecloud-docker ./cmd/rmfakecloud/

FROM scratch
EXPOSE 3000
#ENV RMAPI_HWR_HMAC=""
#ENV RMAPI_HWR_APPLICATIONKEY=""
#ENV RM_SMTP_SERVER=""
#ENV RM_SMTP_USERNAME=""
#ENV RM_SMTP_PASSWORD=""
COPY --from=gobuilder /src/rmfakecloud-docker .
ENTRYPOINT ["/rmfakecloud-docker"]
