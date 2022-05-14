#!/bin/bash
set -o nounset errexit

QSROOT=$(pwd)
export QSROOT
rm -r dist && mkdir dist

# set this for builders behind GFW...
go env -w GOPROXY=https://goproxy.cn,direct
go install github.com/mitchellh/gox@v1.0.1
PATH=$PATH:$HOME/go/bin
cd cmd/start
gox \
    -osarch="linux/amd64" \
    -output "$QSROOT/dist/quickshare/start"

echo "Done"
