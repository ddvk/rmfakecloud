#!/bin/sh
URL="localhost:3000"
curl -d 'email=test&password=test' $URL/ui/api/register
