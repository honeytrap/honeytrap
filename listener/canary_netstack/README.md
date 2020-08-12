# Canary listener

Logs all connections received on specified network interface.


#### Honeytrap configuration for this listener.

iptables rules needed for use with services:

```
iptables -A OUTPUT -p tcp --tcp-flags RST RST -j DROP
iptables -I OUTPUT -p icmp --icmp-type destination-unreachable -j DROP
```


Listener config:
```
[listener]
type="netstack-canary"

# set "iface" to the network interface to use.
interfaces=["iface"]

# block_ports: listed ports are blocked from netstack, traffic will be handled by the host.
# format: "<udp|tcp>/<port-number>"
block_ports=["tcp/22"]

# block_source_ip: blocks all traffic from this ip, blocked traffic is handled by host.
# format: "<ipv4- |ipv6-IP>"
block_source_ip=["1.2.3.4", "2001:0db8:85a3::8a2e:0370:7334"]

# block_destination_ip: blocks all traffic to this ip, blocked traffic is handled by host.
# format: "<ipv4- |ipv6-IP>"
block_destination_ip=[]
```

Warning: if none of the block_* options are used then the honeytrap host is not remote accessible anymore. It's advisible to block the hosts ssh port.

#### TLS:

To be able to accept tls connections a server cerificate- and key-file are required, Thes can be set in the port sections of the honeytrap configuration (config.toml). Different ports can have different certificates. A certificate set on "tcp/0" is used as the default certificate for all tls connections on ports without a certificate.

Example: use one certificate for all tls connections except https on port 443,
```
[[port]]
port="tcp/0"

# certificate_file: file path.
certificate_file="certs/honeytrap.crt"

# key_file: file path.
key_file="certs/honeytrap.key"

[[port]]
port="tcp/443"
services=["http01"]
certificate_file="certs/honeytrap-https.crt"
key_file="certs/honeytrap-https.key"
```

Note 1: A certificate set on port "tcp/0" is required to log all tls connection attempts.

Note 2: Honeytrap services which do there own tls handshake can not be used when a default certificate is set. Instead use the non-tls version, eg. use the http service instead of https.

Note 3: Warnings that there are no services defined on "tcp/0" can be ignored.


#### TODO:

- log more tls data (eg, JA3)
- inbound/outbound option ???? (Now only inbound traffic is logged)
