#!/bin/sh
set -e
echo "Getting the installer..."
INSTALLER="installer.sh"
REPOURL="https://github.com/ddvk/rmfakecloud-proxy/releases/download/v0.0.1/${INSTALLER}"
wget "$REPOURL" -O installer.sh
chmod +x ./${INSTALLER}
echo "Running the installer..."
.${INSTALLER} install
