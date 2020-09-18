#!/bin/sh
if  ! grep rmfake_start /etc/hosts ; then
cat <<EOF >> /etc/hosts
# rmfake_start
127.0.0.1 hwr-production-dot-remarkable-production.appspot.com
127.0.0.1 service-manager-production-dot-remarkable-production.appspot.com
127.0.0.1 local.appspot.com
127.0.0.1 my.remarkable.com
# rmfake_end
EOF
fi

