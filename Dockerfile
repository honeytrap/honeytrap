FROM golang:latest AS builder

ADD . /go/src/github.com/honeytrap/honeytrap
WORKDIR /go/src/github.com/honeytrap/honeytrap

ARG LDFLAGS=""
RUN go build -tags="" -ldflags="$(go run scripts/gen-ldflags.go)" -o /go/bin/app github.com/honeytrap/honeytrap

FROM debian
RUN apt-get update && apt-get install -y ca-certificates curl
COPY --from=builder /go/bin/app /honeytrap/honeytrap

RUN mkdir /config /data

RUN curl -s -o /config/config.toml https://raw.githubusercontent.com/honeytrap/honeytrap-configs/master/server-standalone/config-server-standalone.toml

ENTRYPOINT ["/honeytrap/honeytrap", "--config", "/config/config.toml", "--data", "/data/"]

EXPOSE 8022 5900
