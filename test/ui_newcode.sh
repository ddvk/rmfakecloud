#!/bin/sh
. ./common.env

curl -H "Accept: application/json" -H "Authorization: Bearer $1" $URL/ui/api/newcode
