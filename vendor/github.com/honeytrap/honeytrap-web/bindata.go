package web

//go:generate go-bindata -pkg=web -o=bindata_gen.go -modtime=0 -ignore="index\.bundle\.js\.map" build/...
var Prefix = "build"
