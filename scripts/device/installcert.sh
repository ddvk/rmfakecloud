#!/bin/sh
certdir="/usr/share/ca-certificates"
if [ -f $certdir/ca.crt ];
    echo "The cert has been already installed"
    exit 0
fi
mkdir -p $certdir
cp ca.crt $certdir/
update-ca
