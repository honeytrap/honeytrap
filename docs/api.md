# Honeytrap API

Honeytrap exposes a specific API which allows us to easily retrieve data about sessions and 
events which are occuring with the deployed instance. This API should allow anyone using the project to expose an interface to showcase the different occuring sessions running on the instance.

### HTTP API
The HTTP API exposed by Honeytrap is a `GET` only API which focuses on providing access to 
events and sessions, where the sessions hold's data about containers being used by who and with what credentials and events providing us a view of all processes which occured during the specific container usage and session periods.

- `GET /events`

Expects to recieve a `GET` request to retrieve stored events, with the following body:

```json
{
    "response_per_page": 10,
    "page":1,
    "types": [1,5,20], 
    "sensors": ["ping", "^connect"] 
}
```

Note that all the request body fields are optional and if not present, will instead have all events returned, but if `page` is being used then `response_per_page` must be set as well. 

The `types` and `sensor` field provides a means of filtering based on strings or regular expressions which will filter out the events according to the set criterias.

The API will responds with the following response body as regarding the above request:

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
The `total` field represents the total events records stored within the db.


- `GET /sessions`

Expects to recieve a `GET` request to retrieve stored session events, with the following body:

```json
{
    "response_per_page": 10,
    "page":1,
    "types": [1], 
    "sensors": ["^ssh_"] 
}
```

Note that all the request body fields are optional and if not present, will instead have all events returned, but if `page` is being used then `response_per_page` must be set as well. 

The `types` and `sensor` field provides a means of filtering based on strings or regular expressions which will filter out the events according to the set criterias.

The API will responds with the following response body as regarding the above request:

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

The `total` field represents the total events records stored within the db.

### Websocket API
The Honeytrap exposes also the ability to use websocket to connect with the API to retrieve events and sesssion data, whilst also receiving notifications as to the presence of new events or sessions.

- `GET /ws`
Exposed by the API is a `/ws` route which will attempt to upgrade any http request into a websocket connection which allows interfacing with the API to receive updates:

- Requests
Requests to the API via the websocket endpoint expects requests in JSON format which follow the standard below, this requests are only retrieval and do not store/update any data to the API.

```json
{
 "type": INTEGER value of Request
}
```

The API supports the following requests types with specific int values:

```
FETCH_SESSIONS = 1
FETCH_EVENTS = 3
```

    - `FETCH_SESSIONS` returns all session related events which occur within the system.
    - `FETCH_EVENTS` returns all non-session related events which occur within the system.

- Responses
Resposne from the API via the websocket use the JSON format and follow the layout order:

```json
{
 "type": INTEGER value of Response,
 "payload": JSON Array of Events
}
```

The API supports the following response types with specific int values:

```
FETCH_SESSIONS_REPLY=2
FETCH_EVENTS_REPLY=4
ERROR_RESPONSE = 7
```


- `FETCH_SESSIONS_REPLY` returns all session events retrieved when `FETCH_SESSIONS` request is sent.

- `FETCH_EVENTS_REPLY` returns all session events retrieved when `FETCH_EVENTS` request is sent.

- `ERROR_RESPONSE` is returned if any request sent fails to complete or is rejected due to internal system errors.


Below are response samples:

- `FETCH_SESSIONS`: 

With Request:

```json
{
    "type":1,
}
```

Expected Response if Failed:

```json
{
    "type":7,
    "payload": {
        "request": 1,
        "error": "Failed to retreive events due to db connection"
    }
}
```


Expected Response if Successfully:

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

- `FETCH_EVENTS`:

With Request:

```json
{
    "type":3,
}
```

Expected Response if Failed:

```json
{
    "type":7,
    "payload": {
        "request": 1,
        "error": "Failed to retreive events due to db connection"
    }
}
```


Expected Response if Successfully:

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


The websocket API also provides specific response which contain updates on sessions and non-session events:

```
NEW_SESSIONS=5
NEW_EVENTS=6
```

    - `NEW_SESSIONS` indicates new session events from the backend.

    - `NEW_EVENTS` indicates new non-session events from the backend.

- `NEW_SESSIONS`:

Expected Response Body

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


- `NEW_EVENTS`:

Expected Response Body

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