# Honeytrap [![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/honeytrap/honeytrap?utm_source=badge&utm_medium=badge&utm_campaign=&utm_campaign=pr-badge&utm_content=badge) [![Go Report Card](https://goreportcard.com/badge/honeytrap/honeytrap)](https://goreportcard.com/report/honeytrap/honeytrap) [![Build Status](https://travis-ci.org/honeytrap/honeytrap.svg?branch=master)](https://travis-ci.org/honeytrap/honeytrap)

## What is Honeytrap?
Honeytrap is a honeypot framework written in Go, that isolates each attacker in it's own LXC container. On subsequent attacks, the attacker will be presented with the same container, making monitoring their actions easier. The container events and user sessions can be monitored with an HTTP and WebSocket API. Notifications can be sent to Slack channels. 

## Installation from source

First, install the libraries libpcap-dev for network monitoring, and lxc1 and lxc-dev for container services. 
```
apt install -y libpcap-dev lxc1 lxc-dev
```

Honeytrap is written in Go, so we download the Go language from Google.
```
cd /usr/local
wget https://storage.googleapis.com/golang/go1.8rc3.linux-amd64.tar.gz
tar vxf go1.8rc3.linux-amd64.tar.gz
```

Create a directory for the Honeytrap installation.
```
mkdir /opt/honeytrap
```

Set the Go environment variables for your shell. Add the following to *~/.bashrc*.
```
export GOPATH=/opt/honeytrap
export PATH=$PATH:/usr/local/go/bin/
```

And apply these changes:
```
source ~/.bashrc
```

Now, let's download the application.
```
cd /opt/honeytrap/
go get github.com/honeytrap/honeytrap
```

Copy the sample configuration file for usage.
```
cp ./src/github.com/honeytrap/honeytrap/config.toml.sample /opt/honeytrap/config.toml
```
Now the configuration file will be used automatically. 
Start Honeytrap with the following command:
```
$GOPATH/bin/honeytrap

```
Create a LXC container base image and start it.
```
$ lxc-create -t download -n honeytrap -- --dist ubuntu --release xenial --arch amd64
lxc-start -n honeytrap
```

## API
Honeytrap exposes a specific API which allows us to easily retrieve data about sessions and events which are occurring within the deployed instance. This API allows anyone using the project to expose an interface to showcase the different occurring sessions running on the instance.

### HTTP API
The HTTP API exposed by Honeytrap is a *GET* only API that focuses on providing access to **events** and **sessions**. The sessions contain data about the users and credentials in containers, and events data provides a view of all processes that executed during the specific container usage and session periods.

#### Events
The syntax to receive events information is:
- `GET /events`

This is a `GET` request to retrieve all stored events. Optionally a request body is added, such as the following:


```json
{
    "response_per_page": 10,
    "page":1,
    "types": [1,5,20], 
    "sensors": ["ping", "^connect"] 
}
```

> All the fields in the request body are optional and when ommitted, all events are simply returned. If the `page` field is used, then the `response_per_page` field is also mandatory. The `types` and `sensor` field provide a means of filtering based on strings or regular expressions, filtering out the events based on the set criteria.

The following response body is an example reply to the command above:

```json
{
    "response_per_page": 10,
    "page":1,
    "total":100,
    "events":[
        {
            "type": 1,
            "sensor":"ping",
            "date":"06-04-2013",
            "started":"06-04-2013 01:11:10:32",
            "ended":"06-04-2013 12:11:10:32",
            "token":"43354-57-76767-6767-676334-4343-44334",
            "location":"unknown",
            "category":"connections",
            "hostAddr":"10.78.54.100:7080",
            "localAddr":"43.65.78.2:5000",
            "data":"=b534sfsds34343wwe3443;43434-4343",
            "details": {"extra_data":[]},
            "session_id": "6575-232-4545-232443-55454",
            "container_id": "4343434-43-3434-43434343"
        },
        {
            "type": 1,
            "sensor":"ping",
            "date":"06-04-2013",
            "started":"06-04-2013 01:11:10:32",
            "ended":"06-04-2013 12:11:10:32",
            "token":"43354-57-76767-6767-676334-4343-44334",
            "location":"unknown",
            "category":"connections",
            "hostAddr":"10.78.54.100:7080",
            "localAddr":"43.65.78.2:5000",
            "data":"=b534sfsds34343wwe3443;43434-4343",
            "details": {"extra_data":[]},
            "session_id": "6575-232-4545-232443-55454",
            "container_id": "4343434-43-3434-43434343"
        }
    ]
}
```
The `total` field represents the total events records stored within the database.

#### Sessions
The syntax to receive session information is:
- `GET /sessions`

This is a `GET` request to retrieve all stored session data. Optionally a request body is added, such as the following:

```json
{
    "response_per_page": 10,
    "page":1,
    "types": [1], 
    "sensors": ["^ssh_"] 
}
```

> Note that as with the events reqest, All the fields in the request body are optional and when ommitted, all events are simply returned. If the `page` field is used, then the `response_per_page` field is also mandatory.  The `types` and `sensor` field provide a means of filtering based on strings or regular expressions, filtering out the events based on the set criteria.

The following response body is an example reply to the command above:

```json
{
    "response_per_page": 10,
    "page":1,
    "total":100, 
    "events":[
        {
            "type": 1,
            "sensor":"ssh_session",
            "date":"06-04-2013",
            "started":"06-04-2013 01:11:10:32",
            "ended":"06-04-2013 12:11:10:32",
            "token":"43354-57-76767-6767-676334-4343-44334",
            "location":"unknown",
            "category":"SSHConnections",
            "hostAddr":"10.78.54.100:7080",
            "localAddr":"43.65.78.2:5000",
            "data":"=b534sfsds34343wwe3443;43434-4343",
            "details": {"extra_data":[]},
            "session_id": "6575-232-4545-232443-55454",
            "container_id": "4343434-43-3434-43434343"
        },
    ]
}
```

The `total` field represents the total session records stored within the database.

### WebSocket API
Honeytrap is also able to use WebSockets to connect to the API to retrieve events and session data, and receiving notifications when new events or sessions are detected.

- `GET /ws`
The exposed `/ws` route will attempt to upgrade any HTTP request to a WebSocket connection which allows interfacing with the API to receive updates.

#### Requests
Requests to the API via the WebSocket endpoint are expected in JSON format seen below. These requests only retrieve data and do not store or update any data through the API.

```json
{
 "type": INTEGER value of Request
}
```

The API supports the following request types with specific integer values:

```
FETCH_SESSIONS = 1
FETCH_EVENTS = 3
```

- `FETCH_SESSIONS` returns all session related events that occur within the system.
- `FETCH_EVENTS` returns all non-session related events that occur within the system.

#### Responses
Responses from the API via the WebSocket are in the JSON format and use the following order:

```json
{
 "type": INTEGER value of Response,
 "payload": JSON Array of Events
}
```

The API supports the following response types with specific integer values:

```
FETCH_SESSIONS_REPLY=2
FETCH_EVENTS_REPLY=4
ERROR_RESPONSE = 7
```


- `FETCH_SESSIONS_REPLY` returns all session events when `FETCH_SESSIONS` request is sent.

- `FETCH_EVENTS_REPLY` returns all session events when `FETCH_EVENTS` request is sent.

- an `ERROR_RESPONSE` is returned if any request sent fails to complete or is rejected due to internal system errors.

##### Example Responses
To clarify what happens, some example requests and response examples are provided in this section.

Request with request body:
`FETCH_SESSIONS`
```json
{
    "type": 1,
}
```

The expected response, if failed:

```json
{
    "type":7,
    "payload": {
        "request": 1,
        "error": "Failed to retreive events due to db connection"
    }
}
```


The expected response when successful:

```json
{
    "type": 2,
    "payload":[
        {
            "type": 1,
            "sensor":"ssh_session",
            "date":"06-04-2013",
            "started":"06-04-2013 01:11:10:32",
            "ended":"06-04-2013 12:11:10:32",
            "token":"43354-57-76767-6767-676334-4343-44334",
            "location":"unknown",
            "category":"SSHConnections",
            "hostAddr":"10.78.54.100:7080",
            "localAddr":"43.65.78.2:5000",
            "data":"=b534sfsds34343wwe3443;43434-4343",
            "details": {"extra_data":[]},
            "session_id": "6575-232-4545-232443-55454",
            "container_id": "4343434-43-3434-43434343"
        },
    ]
}
```

Another example request with request body, this time for `FETCH_EVENTS`:

```json
{
    "type": 3,
}
```

The expected response, if failed:

```json
{
    "type":7,
    "payload": {
        "request": 1,
        "error": "Failed to retreive events due to db connection"
    }
}
```


The expected response when successful:

```json
{
    "type": 4,
    "payload":[
        {
            "type": 1,
            "sensor":"ping",
            "date":"06-04-2013",
            "started":"06-04-2013 01:11:10:32",
            "ended":"06-04-2013 12:11:10:32",
            "token":"43354-57-76767-6767-676334-4343-44334",
            "location":"unknown",
            "category":"connections",
            "hostAddr":"10.78.54.100:7080",
            "localAddr":"43.65.78.2:5000",
            "data":"=b534sfsds34343wwe3443;43434-4343",
            "details": {"extra_data":[]},
            "session_id": "6575-232-4545-232443-55454",
            "container_id": "4343434-43-3434-43434343"
        },
    ]
}
```

### Updating Events and Sessions
The WebSocket API also provides a specific response which contains updates for sessions and non-session events. `NEW_SESSIONS` indicate new session events from the backend and `NEW_EVENTS` indicate new non-session events from the backend.

```
NEW_SESSIONS=5
NEW_EVENTS=6
```

#### Examples
When requesting a new session with`NEW_SESSIONS`, the expected response body is:

```json
{
    "type": 6,
    "payload":[
        {
            "type": 1,
            "sensor":"ssh_session",
            "date":"06-04-2013",
            "started":"06-04-2013 01:11:10:32",
            "ended":"06-04-2013 12:11:10:32",
            "token":"43354-57-76767-6767-676334-4343-44334",
            "location":"unknown",
            "category":"SSHConnections",
            "hostAddr":"10.78.54.100:7080",
            "localAddr":"43.65.78.2:5000",
            "data":"=b534sfsds34343wwe3443;43434-4343",
            "details": {"extra_data":[]},
            "session_id": "6575-232-4545-232443-55454",
            "container_id": "4343434-43-3434-43434343"
        },
    ]
}
```

When requesting a new event with `NEW_EVENTS`, the expected response body is:

```json
{
    "type": 5,
    "payload":[
        {
            "type": 1,
            "sensor":"ping",
            "date":"06-04-2013",
            "started":"06-04-2013 01:11:10:32",
            "ended":"06-04-2013 12:11:10:32",
            "token":"43354-57-76767-6767-676334-4343-44334",
            "location":"unknown",
            "category":"connections",
            "hostAddr":"10.78.54.100:7080",
            "localAddr":"43.65.78.2:5000",
            "data":"=b534sfsds34343wwe3443;43434-4343",
            "details": {"extra_data":[]},
            "session_id": "6575-232-4545-232443-55454",
            "container_id": "4343434-43-3434-43434343"
        },
    ]
}
```


## Contribute

Contributions are welcome.

### Setup your Honeytrap Github Repository

Fork Honeytrap upstream source repository to your own personal repository. Copy the URL for marija from your personal github repo (you will need it for the git clone command below).

```sh
$ mkdir -p $GOPATH/src/github.com/honeytrap/honeytrap
$ cd $GOPATH/src/github.com/honeytrap/honeytrap
$ git clone <paste saved URL for personal forked honeytrap repo>
$ cd honeytrap/honeytrap
```

###  Developer Guidelines
``Honeytrap`` community welcomes your contribution. To make the process as seamless as possible, we ask for the following:
* Go ahead and fork the project and make your changes. We encourage pull requests to discuss code changes.
    - Fork it
    - Create your feature branch (git checkout -b my-new-feature)
    - Commit your changes (git commit -am 'Add some feature')
    - Push to the branch (git push origin my-new-feature)
    - Create new Pull Request

* If you have additional dependencies for ``Honeytrap``, ``Honeytrap`` manages its dependencies using [govendor](https://github.com/kardianos/govendor):
    - Run `go get foo/bar`.
    - Edit your code to import foo/bar.
    - Run `make pkg-add PKG=foo/bar` from the top-level directory.

* If you have dependencies for ``Honeytrap`` which needs to be removed:
    - Edit your code to not import foo/bar.
    - Run `make pkg-remove PKG=foo/bar` from top-level directory.

* When you're ready to create a pull request, be sure to:
    - Have test cases for the new code. If you have questions about how to do it, please ask in your pull request.
    - Run `make verifiers`
    - Squash your commits into a single commit. `git rebase -i`. It's okay to force-update your pull request.
    - Make sure `go test -race ./...` and `go build` completes.

* Read [Effective Go](https://github.com/golang/go/wiki/CodeReviewComments) article from Golang project.
    - `Honeytrap` project fully conforms to Golang style.
    - If you found offending code, please feel free to send a pull request.

## Creators

**Remco Verhoef**
- <https://twitter.com/remco_verhoef>
- <https://twitter.com/dutchcoders>

## Copyright and license

Code and documentation copyright 2017 Honeytrap.

Code released under [Affero General Public License](LICENSE).
