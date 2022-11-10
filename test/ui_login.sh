#!/bin/sh
. ./common.env

curl --cookie-jar $COOKIES -sS -d '{"email":"'$USER'", "password":"'$PASS'"}' $URL/ui/api/login
