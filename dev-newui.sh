#!/bin/bash

# This starts the webpack devserver and proxies the api requests to the backend
dir=$(dirname $0)
pushd $dir
export RM_SMTP_SERVER=localhost:2525
export RM_SMTP_NOTLS=1
export JWT_SECRET_KEY=dev
export LOGLEVEL=${1:-DEBUG}
export STORAGE_URL=http://$(hostname):3000
make run-newui &
PID=$!
trap "kill $PID" EXIT
find . -path new-ui -prune -false -o -iname "*.go" | entr -r make new-run
popd
