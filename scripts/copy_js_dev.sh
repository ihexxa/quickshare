#!/bin/bash
set -o nounset errexit

export ROOT=`pwd`
cp $ROOT/node_modules/immutable/dist/immutable.min.js $ROOT/static/public/js/
cp $ROOT/node_modules/react-dom/index.js $ROOT/static/public/js/react-dom.development.js
cp $ROOT/node_modules/react/jsx-dev-runtime.js $ROOT/static/public/js/react.development.js
