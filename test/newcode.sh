#!/bin/sh
URL="localhost:3000"
curl -H "Accept: application/json" -H "Authorization: Bearer $1" $URL/ui/api/newcode
