#!/bin/sh
. ./common.env

./ui_login.sh > /dev/null
curl -sS -b $COOKIES \
    -X GET \
    $URL/ui/api/documents
