# Honeytrap [![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/honeytrap/honeytrap?utm_source=badge&utm_medium=badge&utm_campaign=&utm_campaign=pr-badge&utm_content=badge) [![Go Report Card](https://goreportcard.com/badge/honeytrap/honeytrap)](https://goreportcard.com/report/honeytrap/honeytrap) [![Build Status](https://travis-ci.org/honeytrap/honeytrap.svg?branch=master)](https://travis-ci.org/honeytrap/honeytrap)

----
Honeytrap is an extensible and opensource system for running, monitoring and managing honeypots. 
----

Using honeytrap you can create custom honeypots, consisting of simple services, containers and real hosts combined. Every event will be recorded and send to the configured event channels. 

Usages:

* sinkholing
* canaries

## Platforms

Honeytrap will run on several platforms and architectures. Depending on the platform and architecture functionality will be available.

* Linux (amd64, i386 and arm)
* MacOS
* FreeBSD
* Windows

## Arguments

Argument | Description | Value | Default
--- | --- | --- | ---
--help | show help | | 
--version | show version | | 
--cpu-profiler | run with cpu profiler | | 
--mem-profiler | run with memory profiler | | 
--profiler | start profiler web handler | | 
--list-services | enumerate the available services | |
--list-listeners | enumerate the available listeners | | 
--list-channels | enumerate the available channels | |
--config {file}| use configuration from file | | config.toml

# Development

If you want to write your own listener, director or event channel, you'll need to start here.

## Compilation

Honeytrap compiles on several platforms, depending on the support for the platform functionality is being enabled. As an example, the high interaction honeypot LXC runs only under Linux. 

Make sure the GOPATH is correctly and run this command:

```sh
go build -o "bin/honeytrap-serve-linux-amd64" -ldflags "$LDFLAGS" cmd/honeytrap-serve/main.go; and bin/honeytrap-serve-linux-amd64
```

## Components

### Web

The web interface is being used as a dashboard, but also for configuration. Here you can enable responders, events etc. 

### Listeners

* Socket(socket): this is the network listener for specific ports
* Raw(raw): this listener will listen for all traffic on all ports
* TAP (Linux) (not implemented yet)
* TUN (MacOS, Linux) (not implemented yet)

```
https://serverfault.com/questions/523236/how-do-i-forward-nat-all-traffic-to-one-interface-ip-to-a-remote-ip
```
```
1) Make Server A the next hop for Server B for the traffic in question, which is why it works for your router as mentioned. This could be accomplished, in order of cludgeyness, by making server A the default route for Server B, or using policy routing, or using some fancy iptables, or using a tunnel of some sort.
```

```
sudo ifconfig utun2 10.1.0.10 10.1.0.20 up 
```

``` 
add to tap/ tun to bridge
http://brezular.com/2011/06/19/bridging-qemu-image-to-the-real-network-using-tap-interface/
```

### Directors

* LXC(lxc): director for containing traffic into a personalized lxc container
* Remote(remote): will forward the traffic to a remote host
* Qemu(qemu): will start and forward traffic to qemu machines (not implemented yet)

## Services

* HTTP(http)
* cifs
* webdav
* email
* http image
* 9200 elasticsearch

### Channels 

Events can be send to several channels, to be configured in the configuration file. All channels can be filtered, for example you'll be able to filter specific messages to be sent to Slack, others to Elasticsearch.

* Dummy: this is just a dummy channel and will be used as default
* Slack: send events into Slack channels
* Elasticsearch: send events to Elasticsearch index
* Console: output events to console
* File: output events to file
* Honeyhive: output events to Honeyhive
* Kafka: put events on kafka queue

## Installation

## Install Go 

```sh
cd /usr/local
wget https://storage.googleapis.com/golang/go1.9.linux-amd64.tar.gz
tar vxf go1.9.linux-amd64.tar.gz
```

## Installation from source


```sh
apt install -y libpcap-dev lxc-dev

mkdir /opt/honeytrap
cd /opt/honeytrap/

export GOPATH=/opt/honeytrap
export PATH=$PATH:/usr/local/go/bin/

go get github.com/honeytrap/honeytrap/...

cp config.toml.sample config.toml
$GOPATH/bin/honeytrap
```

## Create Honeytrap template

If you want to run the high interaction container, you need to setup a base image to be used as container template.

```sh
lxc-create -t download -n honeytrap -- --dist ubuntu --release trusty --arch amd64
```

## Contribute to Honeytrap

Please follow [Honeytrap Contributor's Guide](CONTRIBUTING.md)

## Creators

[insert dutchsec logo here]

DutchSec’s mission is to safeguard the evolution of technology and therewith humanity. By delivering  groundbreaking and solid, yet affordable security solutions we make sure no people, companies or institutes are harmed while using technology. We aim to make cyber security available for everyone.

Our team consists of boundary pushing cyber crime experts, grey hat hackers and developers specialized in big data, machine learning, data- and context driven security. By building open source and custom-made security tooling we protect and defend data, both offensively and proactively. 

We work on the front line of security development and explore undiscovered grounds to fulfill our social (and corporate) responsibility. We are driven by the power of shared knowledge and constant learning, and hope to instigate critical thinking in all who use technology in order to increase worldwide safety. We therefore stimulate an open culture, without competition or rivalry, for our own team, as well as our clients.Security is what we do, safety is what you get.

## Copyright and license

Code and documentation copyright 2017 DutchSec.

Code released under [Affero General Public License](LICENSE).
