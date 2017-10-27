# Honeytrap [![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/honeytrap/honeytrap?utm_source=badge&utm_medium=badge&utm_campaign=&utm_campaign=pr-badge&utm_content=badge) [![Go Report Card](https://goreportcard.com/badge/honeytrap/honeytrap)](https://goreportcard.com/report/honeytrap/honeytrap) [![Build Status](https://travis-ci.org/honeytrap/honeytrap.svg?branch=master)](https://travis-ci.org/honeytrap/honeytrap)

----
Honeytrap is an extensible and opensource system for running, monitoring and managing honeypots. 
----

Using honeytrap you can create custom honeypots, consisting of simple services, containers and real hosts.

Honeytrap has three modes, sensor mode, high- and low interaction mode. The sensor mode just detects traffic, this will be ideally used for detection of movement within your network. Low interaction mode will reply with predefined default responses to requests, following playbooks. High interaction honeypots will be 

Usages:

* sinkholing

* directors: this will define the functionality 
* listeners: mode to listen, this can be specific port, raw or using tap/tun interface
* channels: how to send events (elasticsearch, splunk, slack, web interface)

## Sensor
Sensor listens on all ports and receives payloads.

```
$ honeytrap sensor
```

## Low Interaction
Low Interaction listen on specific ports and are being handled by specific protocol implementations. SCADA

```
$ honeytrap low-interaction
```

## High Interaction
High Interaction spawns a container per attacker.

```
$ honeytrap high-interaction
```

## Platforms
The following platforms are supported:

* Linux
* Mac OS
* Windows

## Arguments

Argument | Description | Value
--- | --- | ---
i | interface | 

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

## Contribute

Contributions are welcome.

### Setup your Honeytrap Github Repository

Fork Honeytrap upstream source repository to your own personal repository. Copy the URL for marija from your personal github repo (you will need it for the git clone command below).

```sh
mkdir -p $GOPATH/src/github.com/honeytrap/honeytrap
cd $GOPATH/src/github.com/honeytrap/honeytrap
git clone <paste saved URL for personal forked honeytrap repo>
cd honeytrap/honeytrap
```

###  Developer Guidelines
``Honeytrap`` community welcomes your contribution. To make the process as seamless as possible, we ask for the following:
* Go ahead and fork the project and make your changes. We encourage pull requests to discuss code changes.
    - Fork it
    - Create your feature branch (git checkout -b my-new-feature)
    - Commit your changes (git commit -am 'Add some feature')
    - Push to the branch (git push origin my-new-feature)
    - Create new Pull Request

* If you have additional dependencies for ``Honeytrap``, ``Honeytrap`` manages its dependencies using [govendor](https://github.com/kardianos/govendor)
    - Run `go get foo/bar`
    - Edit your code to import foo/bar
    - Run `make pkg-add PKG=foo/bar` from top-level directory

* If you have dependencies for ``Honeytrap`` which needs to be removed
    - Edit your code to not import foo/bar
    - Run `make pkg-remove PKG=foo/bar` from top-level directory

* When you're ready to create a pull request, be sure to:
    - Have test cases for the new code. If you have questions about how to do it, please ask in your pull request.
    - Run `make verifiers`
    - Squash your commits into a single commit. `git rebase -i`. It's okay to force update your pull request.
    - Make sure `go test -race ./...` and `go build` completes.

* Read [Effective Go](https://github.com/golang/go/wiki/CodeReviewComments) article from Golang project
    - `Honeytrap` project is fully conformant with Golang style
    - if you happen to observe offending code, please feel free to send a pull request

## Creators

[insert dutchsec logo here]

DutchSec’s mission is to safeguard the evolution of technology and therewith humanity. By delivering  groundbreaking and solid, yet affordable security solutions we make sure no people, companies or institutes are harmed while using technology. We aim to make cyber security available for everyone.

Our team consists of boundary pushing cyber crime experts, grey hat hackers and developers specialized in big data, machine learning, data- and context driven security. By building open source and custom-made security tooling we protect and defend data, both offensively and proactively. 

We work on the front line of security development and explore undiscovered grounds to fulfill our social (and corporate) responsibility. We are driven by the power of shared knowledge and constant learning, and hope to instigate critical thinking in all who use technology in order to increase worldwide safety. We therefore stimulate an open culture, without competition or rivalry, for our own team, as well as our clients.Security is what we do, safety is what you get.

## Copyright and license

Code and documentation copyright 2017 DutchSec.

Code released under [Affero General Public License](LICENSE).
