#!/bin/sh
set -e
cd `dirname $0`

. ./common.env

if [ ! -f ./device.token ]; then
    CODE=`./ui_getcode.sh | ./connect.sh`
    echo $CODE > ./device.token
fi

CODE=$(cat ./device.token)
curl -sS -H "Authorization: Bearer $CODE" -X POST $URL/token/json/2/user/new
