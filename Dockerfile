FROM influx6/lxcontains-ubuntu

ADD . /go/src/github.com/honeytrap/honeytrap

RUN bash -c "mkdir -p /honeytrap"
RUN cp config.toml.sample /honeytrap/config.toml

WORKDIR /honeytrap

EXPOSE 8022
EXPOSE 3000

ENTRYPOINT honeytrap
