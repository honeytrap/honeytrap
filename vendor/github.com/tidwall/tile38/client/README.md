Tile38 Client
=============

[![Build Status](https://travis-ci.org/tidwall/tile38.svg?branch=master)](https://travis-ci.org/tidwall/tile38)
[![GoDoc](https://godoc.org/github.com/tidwall/tile38/client?status.svg)](https://godoc.org/github.com/tidwall/tile38/client)

Tile38 Client is a [Go](http://golang.org/) client for [Tile38](http://tile38.com/).

THIS LIBRARY IS DEPRECATED
==========================

Please use the [redigo](https://github.com/garyburd/redigo) client library instead.
If you need JSON output with Redigo then call:
```
conn.Do("OUTPUT", "JSON")
```
