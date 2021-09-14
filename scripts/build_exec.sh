#!/bin/bash
set -o nounset errexit

export QSROOT=`pwd`
rm -r dist && mkdir dist
docker builder prune

# set this for builders behind GFW...
go env -w GOPROXY=https://goproxy.cn,direct
go get -d -v ./...
go get github.com/mitchellh/gox
cd cmd/start
gox \
    -osarch="linux/amd64" \
    -output "$QSROOT/dist/quickshare/start"
