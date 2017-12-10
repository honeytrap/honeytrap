FROM golang:latest

ADD . /go/src/github.com/honeytrap/honeytrap

ARG LDFLAGS=""
RUN mkdir /config/
ADD config-docker.toml /config/config.toml
RUN go build -tags="" -ldflags="$LDFLAGS" -o /go/bin/app github.com/honeytrap/honeytrap

WORKDIR /go/src/github.com/honeytrap/honeytrap

ENTRYPOINT ["/go/bin/app", "--config", "/config/config.toml"]

EXPOSE 8022 5900
