#!/bin/bash
set -o nounset errexit

export QSROOT=`pwd`
rm -r dist && mkdir dist

# set this for builders behind GFW...
go env -w GOPROXY=https://goproxy.cn,direct
go get github.com/mitchellh/gox
cd cmd/start
gox \
    -osarch="windows/386 windows/amd64 darwin/amd64 linux/386 linux/amd64 linux/arm linux/arm64" \
    -output "$QSROOT/dist/quickshare_{{.OS}}_{{.Arch}}/quickshare"

distributions=('quickshare_windows_386' 'quickshare_windows_amd64' 'quickshare_darwin_amd64' 'quickshare_linux_386' 'quickshare_linux_amd64' 'quickshare_linux_arm' 'quickshare_linux_arm64')

cd $QSROOT
for dist in ${distributions[@]}
do
    cp -R $QSROOT/public $QSROOT/dist/$dist # $QSROOT/public must be ready
    cp $QSROOT/configs/lan.yml $QSROOT/dist/$dist
    zip -r -q $QSROOT/dist/$dist.zip ./dist/$dist/*
    rm -r $QSROOT/dist/$dist
done

echo "Done"