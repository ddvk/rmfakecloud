#!/bin/sh

#gets code for dev
token=`curl -s -H "Content-Type: application/json" -d'{"email":"test","password":"test"}' -X POST localhost:3000/ui/api/login`
code=`curl -s -H"Authorization: Bearer $token" localhost:3000/ui/api/newcode | tr -d '\"'` 
echo $code

