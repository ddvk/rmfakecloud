#!/bin/sh
. ./common.env

./ui_login.sh > /dev/null

FOLDER=$1
PARENT=$2
curl -sS -b $COOKIES \
    -X POST \
    -d '{"parentId":"'$PARENT'","name":"'$FOLDER'"}' \
    $URL/ui/api/folders
