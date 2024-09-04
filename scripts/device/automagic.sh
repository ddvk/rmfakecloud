#!/bin/sh
set -e
echo "Getting the installer..."
INSTALLER="installer.sh"
REPOURL="https://github.com/ddvk/rmfakecloud-proxy/releases/download/v0.0.4/${INSTALLER}"
wget "$REPOURL" -O ${INSTALLER}
chmod +x ./${INSTALLER}
echo "Running the installer..."
./${INSTALLER} install
