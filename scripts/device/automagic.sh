#!/bin/sh
REPOURL="https://github.com/ddvk/rmfakeclout/raw/master/scripts/device/"

# workdir
# get stuff
# wget
echo "Getting assets..."
assets=(secure fixsync.sh gencert.sh patchhosts.sh installcert.sh installproxy.sh )
for app in "${assets[@]}"
do
    wget "$REPOURL/$app" -O $app
done
# gencert
echo "Generating certificates..."
./gencert.sh
# installcert
echo "Installing the root ca"
./installcert.sh

read -p "Enter your own cloud url: " url
# install proxy
echo "Proxy set to point to $url"
./installproxy.sh $url
echo "Overriding default cloud addresses..."
./patchhosts.sh

echo "Stopping xochitl and marking all files as not synced"
systemctl stop xochitl
# fix sync
./fixsync.sh
systemctl start xochitl

echo "Done"
