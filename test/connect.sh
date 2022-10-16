#!/bin/sh
set -e
cd `dirname $0`

if [ $# -lt 1 ]; then
    >&2 echo "missing code arg, read from stdin"
    read CODE
else
    CODE=$1
fi
. ./common.env
curl -sS -d '{"code":"'$CODE'", "deviceDesc":"desc", "deviceID":"rm100-123"}' $URL/token/json/2/device/new
