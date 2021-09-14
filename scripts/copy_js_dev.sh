#!/bin/bash
set -o nounset errexit

export ROOT=`pwd`
yarn
cp $ROOT/node_modules/immutable/dist/immutable.min.js $ROOT/public/static/js/
cp $ROOT/node_modules/react-dom/umd/react-dom.development.js $ROOT/public/static/js/
cp $ROOT/node_modules/react/umd/react.development.js $ROOT/public/static/js/
