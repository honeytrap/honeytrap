FROM golang:latest AS go

ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

ADD . /src/honeytrap

ARG LDFLAGS=""

WORKDIR /src/honeytrap
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -tags="" -ldflags="$(go run scripts/gen-ldflags.go)" -o /go/bin/app .

RUN mkdir /config /data
RUN curl -s -o /config/config.toml https://raw.githubusercontent.com/honeytrap/honeytrap-configs/master/server-standalone/config-server-standalone.toml

FROM scratch
MAINTAINER  Remco Verhoef <remco.verhoef@dutchsec.com>

COPY --from=go /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=go /go/bin/app /honeytrap/honeytrap
COPY --from=go /config /config
COPY --from=go /data /data

ENTRYPOINT ["/honeytrap/honeytrap", "--config", "/config/config.toml", "--data", "/data/"]

EXPOSE 8022 5900
