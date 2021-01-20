#!/bin/sh
HOSTPORT=3000
DATA=$(realpath data)
id=$(docker run -d -p $HOSTPORT:3000 -v $DATA:/data -e STORAGE_URL=http://$(hostname):$HOSTPORT --rm rmfakecloud)
echo $id
docker logs $id
