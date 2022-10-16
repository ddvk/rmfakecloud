#!/bin/sh
set -e
cd `dirname $0`

. ./common.env

RECIPIENT=${1:-test@blah}

TOKEN=$(./getusertoken.sh)

curl -F attachment=@test.pdf \
    -F 'from=test@test.com' \
    -F 'to='$RECIPIENT'' \
    -F 'html=blahblah' \
    -H "Authorization: Bearer $TOKEN" \
    -sS \
    -X POST $URL/api/v2/document 
