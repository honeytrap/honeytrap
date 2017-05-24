# Honeytrap Configuration

Honeytrap instances are configured through the use of a `toml` file, which defines the different configuration values and channel requirements for each honeytrap instance.

```toml
director = "lxc"
template = "honeytrap"
token = "433UI-56JK-3433NJ-KI954"

[[services]]
service = "http"
enabled = "true"
port = ":8080"

[[services]]
service = "ssh"
port = ":8022"
banner = "SSH-2.0-OpenSSH_6.6.1p1 2020Ubuntu-2ubuntu2"
#key = "./perm/perms"

[backends."api"]
backend = "honeytrap"
host = "http://api.honeytrap.io/"
token = "b0b6e462-ef0b-11e6-abc7-0fb6247f5820"

[backends."backup.file"]
backend = "file"
file = "/store/files/pushers.pub"
ms = "50s"
max_size = 3000

[backends."gophers"]
backend = "slack"
host = "https://hooks.slack.com/services/"
token = "KUL6M39MCM/YU16GBD/VOOW9HG60eDfoFBiMF"

[[channels]]
backends = ["gophers", "backup.file"]
sensors = ["sessions-start", "containers-open"]
categories = ["sessions", "containers", "attackers"]
events = ["new_session", "connection-open", "connection-closed"]

[[channels]]
backends = ["minio", "api"]
categories = ["sessions", "containers"]
sensors = ["sessions-start", "containers-open"]
events = ["connection-open", "connection-closed"]

```

Each section of the config sample above gives a directive to your honeytrap instance. Below is a broader explanation of each piece.

- Container Technology and Security Settings

```toml
director = "lxc"
template = "honeytrap"

token = "433UI-56JK-3433NJ-KI954"
```

This section gives specific settings which tell honeytrap the underline technology to use for creating the honeytrap sessions, generally this is based on the supported options honeytrap has working (Lxc, Cowrie, etc).

It also dictates the security token which a possible client service will require to have access to the running instance (This may change in future for a better implementation).

- Data backends and settings

```
[backends."api"]
backend = "honeytrap"
host = "http://api.honeytrap.io/"
token = "b0b6e462-ef0b-11e6-abc7-0fb6247f5820"

[backends."backup.file"]
backend = "file"
file = "/store/files/pushers.pub"
ms = "50s"
max_size = 3000

[backends."gophers"]
backend = "slack"
host = "https://hooks.slack.com/services/"
token = "KUL6M39MCM/YU16GBD/VOOW9HG60eDfoFBiMF"
```

This section gives directive to honeytrap to configure the provide services as data submission endpoint which allows us to submit events and relevant container data to this backends.

- Data Channels and settings

```
[[channels]]
backends = ["gophers", "backup.file"]
sensors = ["sessions-start", "containers-open"]
categories = ["sessions", "containers", "attackers"]
events = ["new_session", "connection-open", "connection-closed"]

[[channels]]
backends = ["minio", "api"]
categories = ["sessions", "containers"]
sensors = ["sessions-start", "containers-open"]
events = ["connection-open", "connection-closed"]
```

This section defines specifically to honeytrap the categories, sensors and events allowed to be delivered to which given backends, this is the most important section has without this, no backend service will ever receive data from the honeytrap instance. The provide a unique balance about the data and what type of data is delivered to which backend service.

- Protocol settings 

```toml
[[services]]
service = "http"
enabled = "true"
port = ":8080"

[[services]]
service = "ssh"
port = ":8022"
banner = "SSH-2.0-OpenSSH_6.6.1p1 2020Ubuntu-2ubuntu2"
#key = "./perm/perms"
```

This are specific settings for all protocol agents provided by honeytrap for interaction with containerized honeypot sessions, generally this will not be touched by the user, as these defaults are set for such agents.