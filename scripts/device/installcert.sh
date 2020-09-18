#!/bin/sh
certdir="/usr/local/share/ca-certificates"
if [ -f $certdir/ca.crt ]; then
    echo "The cert has been already installed"
    exit 0
fi
mkdir -p $certdir
cp ca.crt $certdir/
update-ca-certificates
