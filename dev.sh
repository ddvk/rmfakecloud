#!/bin/sh

# This starts the webpack devserver and proxies the api requests to the backend
dir=$(dirname $0)
pushd $dir
export JWT_SECRET_KEY=dev
export LOGLEVEL=${1:-DEBUG}
make runui &
PID=$!
trap "kill $PID" EXIT 
find . -path ui -prune -false -o -iname "*.go" | entr -r make run
popd
