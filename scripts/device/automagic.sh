#!/bin/sh
set -e
REPOURL="https://github.com/ddvk/rmfakecloud/raw/master/scripts/device/"

# workdir
if [ -z "$SKIP_DOWNLOAD" ]; then
    systemctl stop proxy || true
    echo "Getting assets..."
    assets=(secure install.sh fixsync.sh gencert.sh patchhosts.sh installcert.sh installproxy.sh cleanup.sh)
    for app in "${assets[@]}"
    do
        wget "$REPOURL/$app" -O $app
        chmod +x $app
    done
fi
./install.sh
