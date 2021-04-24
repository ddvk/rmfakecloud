#!/bin/sh
URL="localhost:3000"
curl -d '{"email":"fake@rmfake", "password":"foobar"}' $URL/ui/api/login
