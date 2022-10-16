#!/bin/sh
. ./common.env

curl -d '{"email":"'$USER'", "password":"'$PASS'"}' $URL/ui/api/login
