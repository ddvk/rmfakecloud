#!/bin/sh
. ./common.env
PARENT=$1
./ui_login.sh > /dev/null
curl -sS -b $COOKIES \
    -X POST \
    -F file=@test.pdf \
    -F parent=$PARENT \
    $URL/ui/api/documents/upload
