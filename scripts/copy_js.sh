#!/bin/bash
set -o nounset errexit

ROOT=$(pwd)
export ROOT

mkdir -p "$ROOT/static/public/js"
cp "$ROOT/node_modules/immutable/dist/immutable.min.js" "$ROOT/static/public/js/"
cp "$ROOT/node_modules/react-dom/umd/react-dom.production.min.js" "$ROOT/static/public/js/"
cp "$ROOT/node_modules/react/umd/react.production.min.js" "$ROOT/static/public/js/"
