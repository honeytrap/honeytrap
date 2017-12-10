#!/bin/bash

DEST=$(mktemp -d)
SRC=$(pwd)

echo "Cloning $SRC into $DEST"

pushd .
cd $DEST
git clone $SRC/.git .

LDFLAGS="$(go run scripts/gen-ldflags.go)"
echo $LDFLAGS

cp $SRC/Dockerfile .
docker build --build-arg LDFLAGS="$LDFLAGS" -t honeytrap:latest .
popd
