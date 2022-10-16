#!/bin/sh
set -e
cd `dirname $0`

. ./common.env

TOKEN=$(./getusertoken.sh)
meta='{"file_name":"test-upload-extension"}'
b64meta=$(echo $meta | base64)
echo $b64meta

curl -d @test.pdf \
    -H 'Rm-Meta: '$b64meta'' \
    -H 'Content-Type: application/pdf' \
    -H "Authorization: Bearer $TOKEN" \
    -sS \
    -X POST $URL/doc/v2/files
