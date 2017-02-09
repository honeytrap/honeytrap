//go:generate protoc --proto_path=$GOPATH/src:$GOPATH/src/github.com/gogo/protobuf/protobuf:. --go_out=. protocol.proto
package protocol
