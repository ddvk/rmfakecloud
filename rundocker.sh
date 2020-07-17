#!/bin/sh
HOSTPORT=3000
DATA=$(realpath data)
docker run -p $HOSTPORT:3000 -v $DATA:/data -e STORAGE_URL=http://$(hostname):$HOSTPORT -it --rm rmfakecloud
