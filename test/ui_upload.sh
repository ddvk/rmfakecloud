#!/bin/sh
. ./common.env

./ui_login.sh > /dev/null
curl -sS -b $COOKIES \
    -X POST \
    -F file=@test.pdf \
    $URL/ui/api/documents/upload
