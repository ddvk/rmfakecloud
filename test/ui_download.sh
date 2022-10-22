#!/bin/sh
. ./common.env

./ui_login.sh > /dev/null

curl -sS -b $COOKIES $URL/ui/api/documents/$1 -o output.pdf
