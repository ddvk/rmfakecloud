#!/bin/sh
echo "Generating certificates..."
./gencert.sh
echo "Installing the root ca"
./installcert.sh

read -p "Enter your own cloud url: " url
# install proxy
./installproxy.sh $url
echo "Overriding default cloud addresses..."
./patchhosts.sh

echo "Stopping xochitl and marking all files as not synced"
systemctl stop xochitl
./fixsync.sh

systemctl start xochitl

echo "Done"
