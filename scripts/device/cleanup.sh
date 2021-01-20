#!/bin/sh
systemctl stop proxy
systemctl disable proxy
rm proxy.key proxy.crt ca.crt ca.srl ca.key proxy.pubkey proxy.csr csr.conf proxy.cfg
rm /usr/local/share/ca-certificates/ca.crt
rm /etc/systemd/system/proxy.service
sed -i '/# rmfake_start/,/# rmfake_end/d' /etc/hosts
echo "Marking files as not synced to prevent data loss"
./fixsync.sh
echo "You can restart xochitl now"
