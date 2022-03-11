#!/bin/bash

# This starts the webpack devserver and proxies the api requests to the backend
dir=$(dirname $0)
pushd $dir
export JWT_SECRET_KEY=dev
export LOGLEVEL=${1:-DEBUG}
export STORAGE_URL=http://$(hostname):3000
make runui &
PID=$!
trap "kill $PID" EXIT 
find . -path ui -prune -false -o -iname "*.go" | entr -r make run
popd
