#!/bin/sh
echo "Setting cloud sync to: $1"
url=$1
workdir=$(dirname $0)
cat > $workdir/proxy.cfg <<EOF
URL=
EOF
cat > /etc/systemd/system/proxy.service <<EOF
[Unit]
Description=reverse proxy
#StartLimitIntervalSec=600
#StartLimitBurst=4
After=home.mount

[Service]
Environment=HOME=/home/root
#EnvironmentFile=$workdir/proxy.cfg
WorkingDirectory=$workdir
ExecStart=$workdir -cert $workdir/proxy.crt -key $workdir/proxy.key ${url}

[Install]
WantedBy=multi-user.target
EOF
systemctl daemon-reload
systemctl enable proxy
systemctl start proxy
