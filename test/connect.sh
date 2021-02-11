#!/bin/sh
URL="localhost:3000"
curl -d '{"code":"'$1'", "deviceDesc":"desc", "deviceID":"rm100-123"}' $URL/token/json/2/device/new
