#!/bin/bash
set -o nounset errexit

export QSROOT
QSROOT=$(pwd)
rm -r dist && mkdir dist
docker builder prune

# set this for users behind GFW...
go env -w GOPROXY=https://goproxy.cn,direct
go get -d -v ./...
go get github.com/mitchellh/gox
cd cmd/start
gox \
    -osarch="linux/amd64" \
    -output "$QSROOT/dist/quickshare/start"
