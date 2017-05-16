# HTTP API
The HTTP API exposed by Honeytrap is a *GET* only API that focuses on providing access to **events** and **sessions**. The sessions contain data about the users and credentials in containers, and events data provides a view of all processes that executed during the specific container usage and session periods.

## Events
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

## Sessions
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

**Further reading**:
- [WebSocket API](WebSocket-API.md) 