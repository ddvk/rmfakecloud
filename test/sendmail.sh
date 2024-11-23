#!/bin/sh
set -e
cd `dirname $0`

. ./common.env

RECIPIENT=${1:-test@blah}

TOKEN=$(./getusertoken.sh)
BODY="blah<script>alert(1)</script>"

curl -F attachment=@test.pdf \
    -F 'from=test@test.com' \
    -F 'to='$RECIPIENT'' \
	-F 'html='$BODY'' \
    -H "Authorization: Bearer $TOKEN" \
    -sS \
    -X POST $URL/api/v2/document 
