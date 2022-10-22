#!/bin/sh
cd `dirname $0`
. ./common.env
#gets code for dev
token=`./ui_login.sh`

code=`curl -sS -b $COOKIES  $URL/ui/api/newcode | tr -d '\"'` 
echo $code

