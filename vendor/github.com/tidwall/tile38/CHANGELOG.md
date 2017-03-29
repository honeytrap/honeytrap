# Change Log
All notable changes to this project will be documented in this file.
This project adheres to [Semantic Versioning](http://semver.org/).

## [1.8.0] - 2017-02-21
### Added
- #145: TCP Keepalives option (@UriHendler)
- #136: K nearest neighbors for NEARBY command (@m1ome, @tomquas, @joernroeder)
- #139: Added CLIENT command (@UriHendler)
- #133: AutoGC config option (@m1ome, @amorskoy)

### Fixed
- #147: Leaking http hook connections (@mkabischev)
- #143: Duplicate data in hook data (@mkabischev)

## [1.7.5] - 2017-01-13
### Added
- Performance bump for all SET commands, ~10% faster
- Lower memory footprint for large datasets
- #112: Added distance to NEARBY command (@m1ome, @auselen)
- #123: Redis endpoint for webhooks (@m1ome)
- #128: Allow disabling HTTP & WebSocket transport (@m1ome)

### Fixed
- #116: Missing response in TTL json command (@phulst)
- #117: Error in command documentation (@juanpabloaj)
- #118: Unexpected EOF bug with websockets (@m1ome)
- #122: Disque typo timeout handling (@m1ome)
- #127: 3d object searches with 2d geojson area (@damariei)

## [1.7.0] - 2016-12-29
### Added
- #104: PDEL command - Selete objects that match a pattern (@GameFreedom)
- #99: COMMAND keyword for masking geofences by command type (@amorskoy)
- #96: SCAN keyword for roaming geofences
- fba34a9: JSET, JGET, JDEL commands

### Fixed
- #107: Memory leak (@amorskoy)
- #98: Output json fix

## [1.6.0] - 2016-12-11
### Added
- #87: Fencing event grouping (@huangpeizhi)

### Fixed
- #91: Wrong winding order for CirclePolygon function (@antonioromano)
- #73: Corruption for AOFSHRINK (@huangpeizhi)
- #71: Lower memory usage. About 25% savings (@thisisaaronland, @umpc)
- Polygon raycast bug. tidwall/poly#1 (@drewlesueur)
- Added black-box testing

## [1.5.4] - 2016-11-17
### Fixed
- #84: Hotfix - roaming fence deadlock (@tomquas)

## [1.5.3] - 2016-11-16
### Added
- #4: Official docker support (@gordysc)

### Fixed
- #77: NX/XX bug (@damariei)
- #76: Match on prefix star (@GameFreedom, @icewukong)
- #82: Allow for precise search for strings (@GameFreedom)
- #83: Faster congruent modulo for points (@icewukong, @umpc)

## [1.5.2] - 2016-10-20
### Fixed
- #70: Invalid results for INTERSECTS query (@thisisaaronland)

## [1.5.1] - 2016-10-19
### Fixed
- #67: Call the EXPIRE command hangs the server (@PapaStifflera)
- #64: Missing points in 'Nearby' queries (@umpc)

## [1.5.0] - 2016-10-03
### Added
- #61: Optimized queries on 3d objects (@damariei)
- #60: Added [NX|XX] keywords to SET command (@damariei)
- #29: Generalized hook interface (@jeremytregunna)
- GRPC geofence hook support 

### Fixed
- #62: Potential Replace Bug Corrupting the Index (@umpc)
- #57: CRLF codes in info after bump from 1.3.0 to 1.4.2 (@olevole)

## [1.4.2] - 2016-08-26
### Fixed
- #49. Allow fragmented pipeline requests (@owaaa)
- #51: Allow multispace delim in native proto (@huangpeizhi)
- #50: MATCH with slashes (@huangpeizhi)
- #43: Linestring nearby search correction (@owaaa)

## [1.4.1] - 2016-08-26
### Added
- #34: Added "BOUNDS key" command (@icewukong)

### Fixed
- #38: Allow for nginx support (@GameFreedom)
- #39: Reset requirepass (@GameFreedom)

## [1.3.0] - 2016-07-22
### Added
- New EXPIRE, PERSISTS, TTL commands. New EX keyword to SET command
- Support for plain strings using `SET ... STRING value.` syntax
- New SEARCH command for finding strings
- Scans can now order descending

### Fixed
- #28: fix windows cli issue (@zhangkaizhao)

## [1.2.0] - 2016-05-24
### Added
- #17: Roaming Geofences for NEARBY command (@ElectroCamel, @davidxv)
- #15: maxmemory config setting (@jrots)

## [1.1.4] - 2016-04-19
### Fixed
- #12: Issue where a newline was being added to HTTP POST requests (@davidxv)
- #13: OBJECT keyword not accepted for WITHIN command (@ray93)
- Panic on missing key for search requests

## [1.1.2] - 2016-04-12
### Fixed
- A glob suffix wildcard can result in extra hits
- The native live geofence sometimes fails connections

## [1.1.0] - 2016-04-02
### Added
- Resp client support. All major programming languages now supported
- Added WITHFIELDS option to GET
- Added OUTPUT command to allow for outputing JSON when using RESP
- Added DETECT option to geofences

### Changes
- New AOF file structure.
- Quicker and safer AOFSHRINK.

### Deprecation Warning
- Native protocol support is being deprecated in a future release in favor of RESP
