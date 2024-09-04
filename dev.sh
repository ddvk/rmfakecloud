#!/bin/bash
set -e

# This starts the webpack devserver and proxies the api requests to the backend
echo "to test SMTP run:" 
echo "     python -m smtpd -n -c DebuggingServer localhost:2525"
echo 
cd $(dirname $0)
export RM_SMTP_SERVER=localhost:2525
export RM_SMTP_NOTLS=1
export RM_SMTP_NOAUTH=1
export JWT_SECRET_KEY=dev
export LOGLEVEL=${1:-DEBUG}
export STORAGE_URL=http://$(hostname):3000
make runui &
PID=$!
trap "kill $PID ||:" EXIT 
find . -path ui -prune -false -o -iname "*.go" | entr -r make run
