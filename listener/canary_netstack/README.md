# Canary listener

Logs all connections received on specified network interface.


#### Honeytrap configuration for this listener.

iptables rules needed for use with services:

```
iptables -A OUTPUT -p tcp --tcp-flags RST RST -j DROP
iptables -I OUTPUT -p icmp --icmp-type destination-unreachable -j DROP
```


```
[listener]
type="netstack-canary"

# set "iface" to the network interface to use.
interfaces=["iface"]

# exclude_log_protos sets the used protocols for logging (optional) (default: all)
# recognized options for protos: ["ip4", "ip6", "arp", "udp", "tcp", "icmp"]
# exclude_log_protos=[]

# no_tls true: checks connection for tls and does tls handshake if so.
# no_tls=false

# certificate and key_file are neccesary for tls connections.
certificate_file="cert.pem"
key_file="key.pem"
```


#### TODO:

- ssh to honeytrap host is not possible.
- inbound/outbound option ???? (Now only inbound traffic is logged)
