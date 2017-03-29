![Tile38 Logo](https://raw.githubusercontent.com/tidwall/tile38/master/doc/logo200.png)

Tile38 is an open source (MIT licensed), in-memory geolocation data store, spatial index, and realtime geofence. It supports a variety of object types including lat/lon points, bounding boxes, XYZ tiles, Geohashes, and GeoJSON. 

*For more information visit: [http://tile38.com](http://tile38.com)*

## Features

- Spatial index with [search](#searching) methods such as Nearby, Within, and Intersects.
- Realtime [geofencing](#geofencing) through persistent sockets or [webhooks](http://tile38.com/commands/sethook).
- Object types of [lat/lon](#latlon-point), [bbox](#bounding-box), [Geohash](#geohash), [GeoJSON](#geojson), [QuadKey](#quadkey), and [XYZ tile](#xyz-tile).
- Support for lots of [Clients Libraries](#client-libraries) written in many different languages.
- Variety of protocols, including [http](#http) (curl), [websockets](#websockets), [telnet](#telnet), and the [Redis RESP](http://redis.io/topics/protocol).
- Server responses are [RESP](http://redis.io/topics/protocol) or [JSON](http://www.json.org).
- Full [command line interface](#cli).
- Leader / follower [replication](#replication).
- In-memory database that persists on disk.

## Contact
Josh Baker [@tidwall](http://twitter.com/tidwall)

## License

Tile38 source code is available under the MIT [License](/LICENSE).
