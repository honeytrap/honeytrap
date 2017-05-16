# WebSocket API
Honeytrap is also able to use WebSockets to connect to the API to retrieve events and session data, and receiving notifications when new events or sessions are detected.

- `GET /ws`
The exposed `/ws` route will attempt to upgrade any HTTP request to a WebSocket connection which allows interfacing with the API to receive updates.

## Requests
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

## Responses
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

###Example Responses
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

# Updating Events and Sessions
The WebSocket API also provides a specific response which contains updates for sessions and non-session events. `NEW_SESSIONS` indicate new session events from the backend and `NEW_EVENTS` indicate new non-session events from the backend.

```
NEW_SESSIONS=5
NEW_EVENTS=6
```

## Examples
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

Further reading:
- [HTTP API](HTTP-API.md) 