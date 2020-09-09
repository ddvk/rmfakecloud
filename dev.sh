#!/bin/sh
make prep
make dev &
PID=$!
trap "kill $PID" EXIT 
make devui
