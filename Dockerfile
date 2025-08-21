FROM golang:1.24-alpine as build

LABEL org.opencontainers.image.source="https://github.com/tyzbit/acars-processor"

WORKDIR /
COPY . ./

RUN apk add \
    build-base \
    git \
    &&  go build -ldflags="-s -w"

FROM alpine

COPY --From=build /acars-processor /

CMD ["/acars-processor"]
