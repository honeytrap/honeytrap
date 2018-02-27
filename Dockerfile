FROM golang:latest AS go

RUN apt update -y

ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

ADD . /go/src/github.com/honeytrap/honeytrap

ARG LDFLAGS=""

WORKDIR /go/src/github.com/honeytrap/honeytrap
RUN go build -tags="" -ldflags="$(go run scripts/gen-ldflags.go)" -o /go/bin/app github.com/honeytrap/honeytrap

FROM debian

COPY --from=go /go/bin/app /honeytrap/honeytrap

RUN mkdir /config/
RUN mkdir /data/
ADD config-docker.toml /config/config.toml

ENTRYPOINT ["/honeytrap/honeytrap", "--config", "/config/config.toml", "--data", "/data/"]

EXPOSE 8022 5900
