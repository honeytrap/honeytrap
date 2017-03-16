FROM influx6/lxcontains-ubuntu

ADD . /go/src/github.com/honeytrap/honeytrap

WORKDIR /go/src/github.com/honeytrap/honeytrap

RUN go get -v

RUN go test -v ./...

RUN go install

RUN bash -c "mkdir -p /honeytrap"
RUN cp config.toml.sample /honeytrap/config.toml
