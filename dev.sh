#!/bin/sh

# This starts the webpack devserver and proxies the api requests to the backend

make prep
make dev &
PID=$!
trap "kill $PID" EXIT 
make devui
