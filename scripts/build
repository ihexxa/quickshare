export QSROOT=`pwd`
rm -r dist && mkdir dist

go get github.com/mitchellh/gox
cd cmd/start
gox \
    -osarch="windows/386 windows/amd64 darwin/amd64 linux/386 linux/amd64 linux/arm linux/arm64" \
    -output "$QSROOT/dist/quickshare_{{.OS}}_{{.Arch}}/quickshare"

cd $QSROOT/src/client/web
yarn build

cd $QSROOT
cp -R $QSROOT/public $QSROOT/dist/quickshare_windows_386
cp -R $QSROOT/public $QSROOT/dist/quickshare_windows_amd64
cp -R $QSROOT/public $QSROOT/dist/quickshare_darwin_amd64
cp -R $QSROOT/public $QSROOT/dist/quickshare_linux_386
cp -R $QSROOT/public $QSROOT/dist/quickshare_linux_amd64
cp -R $QSROOT/public $QSROOT/dist/quickshare_linux_arm
cp -R $QSROOT/public $QSROOT/dist/quickshare_linux_arm64

cd $QSROOT/dist

zip -r -q $QSROOT/dist/quickshare_windows_386.zip quickshare_windows_386/*
zip -r -q $QSROOT/dist/quickshare_windows_amd64.zip quickshare_windows_amd64/*
zip -r -q $QSROOT/dist/quickshare_darwin_amd64.zip quickshare_darwin_amd64/*
zip -r -q $QSROOT/dist/quickshare_linux_386.zip quickshare_linux_386/*
zip -r -q $QSROOT/dist/quickshare_linux_amd64.zip quickshare_linux_amd64/*
zip -r -q $QSROOT/dist/quickshare_linux_arm.zip quickshare_linux_arm/*
zip -r -q $QSROOT/dist/quickshare_linux_arm64.zip quickshare_linux_arm64/*
