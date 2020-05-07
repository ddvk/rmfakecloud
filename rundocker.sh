#!/bin/sh
PORT=3000
DATA=$(realpath data)
docker run -p 3000:$PORT -v $DATA:/data -e STORAGE_URL=http://$(hostname):$PORT -it --rm rmfakecloud
