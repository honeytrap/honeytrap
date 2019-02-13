FROM golang:latest AS go

RUN apt update -y

ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

ADD . /go/src/github.com/honeytrap/honeytrap

ARG LDFLAGS=""

WORKDIR /go/src/github.com/honeytrap/honeytrap
RUN go build -tags="" -ldflags="$(go run scripts/gen-ldflags.go)" -o /go/bin/app github.com/honeytrap/honeytrap

FROM debian

RUN apt-get update && apt-get install -y ca-certificates curl

RUN mkdir /config
RUN mkdir /data

RUN curl -s -o /config/config.toml https://raw.githubusercontent.com/honeytrap/honeytrap-configs/master/server-standalone/config-server-standalone.toml
COPY --from=go /go/bin/app /honeytrap/honeytrap

ENTRYPOINT ["/honeytrap/honeytrap", "--config", "/config/config.toml", "--data", "/data/"]

EXPOSE 8022 5900
