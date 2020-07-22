# Canary listener

Logs all connections received on specified network interface.


#### Honeytrap configuration for this listener.

```
[listener]
type="netstack-canary"

# set "iface" to the network interface to use. (only one interface supported)
interfaces=["iface"]

# exclude_log_protos sets the used protocols for logging (optional) (default: all)
# recognized options for protos: ["ip4", "ip6", "arp", "udp", "tcp", "icmp"]
# exclude_log_protos=[]
```


#### TODO:

- Use honeytrap services with this listener. This is implemented but not working now, needs fix.
- TLS on/off option ????
- inbound/outbound option ???? (Now only inbound traffic is logged)
