#!/bin/sh
. ./common.env

./ui_login.sh > /dev/null

DOCUMENTID=$1

curl -sS -b $COOKIES \
    -X DELETE \
    $URL/ui/api/documents/$DOCUMENTID
