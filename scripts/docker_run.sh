docker run \
    --name quickshare \
    -d -p :8686:8686 \
    -v `pwd`/quickshare/root:/quickshare/root \
    -e DEFAULTADMIN=qs \
    -e DEFAULTADMINPWD=1234 \
    hexxa/quickshare:0.3.0 