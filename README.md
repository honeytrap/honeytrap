# Honeytrap [![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/honeytrap/honeytrap?utm_source=badge&utm_medium=badge&utm_campaign=&utm_campaign=pr-badge&utm_content=badge) [![Go Report Card](https://goreportcard.com/badge/honeytrap/honeytrap)](https://goreportcard.com/report/honeytrap/honeytrap) [![Build Status](https://travis-ci.org/honeytrap/honeytrap.svg?branch=master)](https://travis-ci.org/honeytrap/honeytrap)


<img src="honeytrap_icon-small.png"/>

## What is Honeytrap?
Honeytrap is a honeypot framework written in Go, that isolates each attacker in it's own LXC container. On subsequent attacks, the attacker will be presented with the same container, making monitoring their actions easier. The container-events and user-sessions can be monitored with an HTTP and WebSocket API. Logging can also be sent to other locations like Slack chatrooms. For more information and news, be sure to visit our official website or subscribe to our Twitter feed.

- [Official website](http://honeytrap.io/#!/)
- [Twitter](https://twitter.com/honeycastio)

## Installation
Currently Honeytrap can only be installed from source on Linux, since it depends on Linux Containers (LXC). It has been tested on Linux (CentOS and Ubuntu) and also works on a Raspberry Pi. Our guide is provided [here](https://github.com/Einzelganger/honeytrap/wiki/Installation).
> Note that the Dockerfile is there for autobuilding purposes, not for installing Honeytrap.


## API
Honeytrap exposes an API that makes it easy to retrieve data about sessions and events that occur within the deployed instance. This API allows anyone using the project to expose an interface to showcase the different occurring sessions running on the instance.

More information:
- The **HTTP API** manual can be found [here](https://github.com/Einzelganger/honeytrap/wiki/HTTP-API.md) 
- The  **WebSocket API** manual can be found [here](https://github.com/Einzelganger/honeytrap/wiki/WebSocket-API.md) 


## Contribute
Contributions are [welcome](https://github.com/Einzelganger/honeytrap/wiki/Contribution_Guide).

## About
We hope you enjoy this program. If you have any comments, tips or want to thank us, you can find us here.

**Remco Verhoef**
- <https://twitter.com/remco_verhoef>
- <https://twitter.com/dutchcoders>
