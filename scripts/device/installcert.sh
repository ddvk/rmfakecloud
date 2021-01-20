#!/bin/sh
certdir="/usr/local/share/ca-certificates"
certname=$certdir/ca.crt
if [ -f $certname ]; then
    echo "The cert has been already installed, it will be removed and reinstalled!!!"
    rm  $certname
    update-ca-certificates --fresh
fi
mkdir -p $certdir
cp ca.crt $certdir/
update-ca-certificates --fresh
