Building
========

Dependencies
------------

To be able to compile from source, you'll need the following dependencies:

* [nodejs](https://nodejs.org) version 16 at most
* [yarn](https://yarnpkg.com/)
* [go](https://go.dev/) version 1.16 at least
* make

Build
-----

```sh
git clone https://github.com/ddvk/rmfakecloud
cd rmfakecloud
make all
```

Installing
==========

1. Copy the `rmfakecloud` binary for your system from the `dist` folder to `/usr/local/bin` and rename it to `rmfakecloud`
   e.g. `cp dist/rmfakecloud-x64 /usr/local/bin/rmfakecloud`
   or `scp dist/rmfakecloud-armv7 raspberry:/usr/local/bin/rmfakecloud`
2. Setup the service to run with your init system. See below for examples
3. Create and modify the configuration file. See below for examples
4. Create the library folder you specified in your configuration file.
   e.g. `mkdir /usr/local/lib/rmfakecloud`
5. Enable and start the service with your init system.
   e.g. `rc-update add rmfakecloud && service start rmfakecloud` or `systemctl enable --now rmfakecloud`

Init System Examples
====================

OpenRC
------

/etc/init.d/rmfakecloud

```sh
#!/sbin/openrc-run

name="rmfakecloud"
command="/usr/local/bin/rmfakecloud"
command_args=""
pidfile="/var/run/rmfakecloud.pid"
command_background="yes"
output_log="/var/log/messages"
error_log="/var/log/messages"
depend() {
    need net localmount
}
```

/etc/conf.d/rmfakecloud

```sh
# Basic settings
export JWT_SECRET_KEY=SOME_KEY
export STORAGE_URL=http(s)://host.where.rmfakecloud.is.running
export PORT=80
export DATADIR=/usr/local/lib/rmfakecloud
export LOGLEVEL=info
# uncomment if using TLS
#export PORT=443
#export TLS_KEY=/path/to/somekey
#export TLS_CERT=/path/to/somecert
#export RM_HTTPS_COOKIE=1

# Email
export RM_SMTP_SERVER=smtp.gmail.com:465
export RM_SMTP_USERNAME=MY_EMAIL_ADDRESS
export RM_SMTP_PASSWORD=MY_SMTP_OR_APP_PASSWORD
# Handwriting recognition
export RMAPI_HWR_APPLICATIONKEY=SOME_KEY
export RMAPI_HWR_HMAC=SOME_KEY
```

Make sure to replace `SOME_KEY` by the return of `openssl rand -base64 48`, see [configuration](configuration.md).

If using GMail, ensure you enable 2FA on that Google account, generate a GMail app password (https://myaccount.google.com/u/0/apppasswords), and provide the app password instead of the account password above.

systemd
-------

rmfakecloud.service

```ini
[Unit]
Description=rmfakecloud
After=network-online.target
Wants=network-online.target

[Service]
ExecStart=/usr/local/bin/rmfakecloud
EnvironmentFile=/etc/rmfakecloud.conf

[Install]
WantedBy=multi-user.target

```

rmfakecloud.conf

```sh
JWT_SECRET_KEY=SOME_KEY
STORAGE_URL=http(s)://host.where.rmfakecloud.is.running
PORT=80
DATADIR=/usr/local/lib/rmfakecloud
LOGLEVEL=info
# uncomment if using TLS
#PORT=443
#TLS_KEY=/path/to/somekey
#TLS_CERT=/path/to/somecert
#RM_HTTPS_COOKIE=1

# Email
RM_SMTP_SERVER=smtp.gmail.com:465
RM_SMTP_USERNAME=MY_EMAIL_ADDRESS
RM_SMTP_PASSWORD=MY_SMTP_OR_APP_PASSWORD
# Handwriting recognition
RMAPI_HWR_APPLICATIONKEY=SOME_KEY
RMAPI_HWR_HMAC=SOME_KEY
```

Make sure to replace `SOME_KEY` with the output of `openssl rand -base64 48`, see [configuration](configuration.md).

If using GMail, ensure you enable 2FA on that Google account, generate a GMail app password (https://myaccount.google.com/u/0/apppasswords), and provide the app password instead of the account password above.
