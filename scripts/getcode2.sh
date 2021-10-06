#!/bin/sh

cookies=cookie.txt
#gets code for dev
token=`curl -b $cookies -c $cookies -s -H "Content-Type: application/json" -d'{"email":"test","password":"test"}' -X POST localhost:3000/ui/api/login`
code=`curl -b $cookies -c $cookies -s localhost:3000/ui/api/newcode | tr -d '\"'`
echo $code

