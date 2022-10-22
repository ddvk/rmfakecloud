#!/bin/sh
. ./common.env

./ui_login.sh > /dev/null

DOCUMENTID=$1
NEW_NAME=$2
PARENT=$3

curl -sS -b $COOKIES \
    -X PUT \
    -d '{"parentId":"'$PARENT'","name":"'${NEW_NAME}'", "documentId":"'${DOCUMENTID}'"}' \
    $URL/ui/api/documents
