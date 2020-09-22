#!/bin/sh
certdir="/usr/local/share/ca-certificates"
if [ -f $certdir/ca.crt ]; then
    echo "The cert has been already installed, if it was regenerated it will not work!"
    #todo bin compare the files

    exit 0
fi
mkdir -p $certdir
cp ca.crt $certdir/
update-ca-certificates
