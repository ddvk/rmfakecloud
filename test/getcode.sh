#!/bin/sh
cd `dirname $0`
. ./common.env
#gets code for dev
token=`curl -s -H "Content-Type: application/json" -d'{"email":"test","password":"test"}' -X POST $URL/ui/api/login`
code=`curl -sS -H"Authorization: Bearer $token" $URL/ui/api/newcode | tr -d '\"'` 
echo $code

