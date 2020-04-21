#!/bin/sh
if  ! grep patched /etc/hosts ; then
cat <<EOF >> /etc/hosts
# patched
127.0.0.1 service-manager-production-dot-remarkable-production.appspot.com
127.0.0.1 local.appspot.com
127.0.0.1 my.remarkable.com
EOF
fi

