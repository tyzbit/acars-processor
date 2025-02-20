FROM golang:1.24-alpine as build
WORKDIR /
COPY . ./

RUN apk add \
    build-base \
    git \
&&  go build -ldflags="-s -w"

FROM alpine

COPY --from=build /acars-annotator /

CMD ["/acars-annotator"]