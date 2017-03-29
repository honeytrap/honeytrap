#!/bin/bash

cd $(dirname "${BASH_SOURCE[0]}")
protoc --go_out=plugins=grpc,import_path=hservice:. *.proto
