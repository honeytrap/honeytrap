<img src="http://docs.honeytrap.io/images/logo.png" height="110" />

# Honeytrap [![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/honeytrap/honeytrap?utm_source=badge&utm_medium=badge&utm_campaign=&utm_campaign=pr-badge&utm_content=badge) [![Go Report Card](https://goreportcard.com/badge/honeytrap/honeytrap)](https://goreportcard.com/report/honeytrap/honeytrap) [![Build Status](https://travis-ci.org/honeytrap/honeytrap.svg?branch=master)](https://travis-ci.org/honeytrap/honeytrap) [![codecov](https://codecov.io/gh/honeytrap/honeytrap/branch/master/graph/badge.svg)](https://codecov.io/gh/honeytrap/honeytrap) [![Docker pulls](https://img.shields.io/docker/pulls/honeytrap/honeytrap.svg)](https://hub.docker.com/r/honeytrap/honeytrap/)

### Honeytrap is an extensible and opensource system for running, monitoring and managing honeypots. 

## Features

* Combine multiple services to one honeypot, eg a LAMP server
* Honeytrap Agent will download the configuration from the Honeytrap Server
* Use the Honeytrap Agent to redirect traffic out of the network to a seperate network
* Deploy a large amount agents while having one Honeytrap Server, configuration will be downloaded automatically and logging centralized
* Payload detection to determine which service should handle the request, one port can handle multiple protocols
* Monitor lateral movement within your network with the Sensor listener. The sensor will complete the handshake (in case of tcp), and store the payload
* Create high interaction honeypots using the LXC or remote hosts directors, traffic will be man-in-the-middle proxied, while information will be extracted
* Extend honeytrap with existing honeypots (like cowrie or glutton), while using the logging and listening framework of Honeytrap
* Advanced logging system with filtering and logging to Elasticsearch, Kafka, Splunk, Raven, File or Console
* Services are easily extensible and will extract as much information as possible
* Low- to high interaction Honeypots, where connections will be upgraded seamless to high interaction

## To start using Honeytrap

See our documentation on [docs.honeytrap.io](http://docs.honeytrap.io/docs/home/).

## Community
Join the [honeytrap-users](https://groups.google.com/forum/#!forum/honeytrap-users) mailing list to discuss all things Honeytrap.

## Creators

[DutchSec](https://dutchsec.com)’s mission is to safeguard the evolution of technology and therewith humanity. By delivering  groundbreaking and solid, yet affordable security solutions we make sure no people, companies or institutes are harmed while using technology. We aim to make cyber security available for everyone.

Our team consists of boundary pushing cyber crime experts, grey hat hackers and developers specialized in big data, machine learning, data- and context driven security. By building open source and custom-made security tooling we protect and defend data, both offensively and proactively. 

We work on the front line of security development and explore undiscovered grounds to fulfill our social (and corporate) responsibility. We are driven by the power of shared knowledge and constant learning, and hope to instigate critical thinking in all who use technology in order to increase worldwide safety. We therefore stimulate an open culture, without competition or rivalry, for our own team, as well as our clients.Security is what we do, safety is what you get.

## Copyright and license

Code and documentation copyright 2017 DutchSec.

Code released under [Affero General Public License](LICENSE).
