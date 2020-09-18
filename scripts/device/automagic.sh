#!/bin/sh
set -e
REPOURL="https://github.com/ddvk/rmfakecloud/raw/master/scripts/device/"

# workdir
if [ -z "$SKIP_DOWNLOAD" ]; then
    systemctl stop proxy || true
    echo "Getting assets..."
    assets=(secure fixsync.sh gencert.sh patchhosts.sh installcert.sh installproxy.sh )
    for app in "${assets[@]}"
    do
        wget "$REPOURL/$app" -O $app
        chmod +x $app
    done
fi
# gencert
echo "Generating certificates..."
./gencert.sh
# installcert
echo "Installing the root ca"
./installcert.sh

read -p "Enter your own cloud url: " url
# install proxy
./installproxy.sh $url
echo "Overriding default cloud addresses..."
./patchhosts.sh

echo "Stopping xochitl and marking all files as not synced"
systemctl stop xochitl
# fix sync
./fixsync.sh
systemctl start xochitl

echo "Done"
