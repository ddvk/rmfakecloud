Building
========

Dependencies:
-------------

* [nodejs](https://nodejs.org)
* [yarn](https://yarnpkg.com/)
* [go](https://go.dev/)
* make

Build:
------

`git clone https://github.com/ddvk/rmfakecloud`
`make all`

Running
=======

1. Copy the rmfakecloud binary for your system from the `dist` folder to `/usr/local/bin` and rename it to `rmfakecloud` e.g. `cp dist/rmfakecloud-x64 /usr/local/bin/rmfakecloud`
2. Setup the service to run with your init system. See below for examples
3. Create and modify the configuration file. See below for examples
4. Create the library folder you specified in your configuration file. e.g. `mkdir /usr/local/lib/rmfakecloud`
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
export STORAGE_URL=https://example.com
export PORT=80
export DATADIR=/usr/local/lib/rmfakecloud
export LOGLEVEL=info
export RM_HTTPS_COOKIE=1
# Email
export RM_SMTP_SERVER=smtp.gmail.com:465
export RM_SMTP_USERNAME=MY_EMAIL_ADDRESS
export RM_SMTP_PASSWORD=MY_PASSWORD
# Handwriting recognition
export RMAPI_HWR_APPLICATIONKEY=SOME_KEY
export RMAPI_HWR_HMAC=SOME_KEY
```

SystemD
-------
rmfakecloud.service
```ini
[Unit]
Description=rmfakecloud

[Service]
ExecStart=/usr/local/bin/rmfakecloud
EnvironmentFile=/etc/rmfakecloud.conf

[Install]
WantedBy=multi-user.target
Wants=network-online.target
After=network-online.target
```
rmfakecloud.conf
```sh
JWT_SECRET_KEY=SOME_KEY
STORAGE_URL=https://example.com
PORT=80
DATADIR=/usr/local/lib/rmfakecloud
LOGLEVEL=info
RM_HTTPS_COOKIE=1
# Email
RM_SMTP_SERVER=smtp.gmail.com:465
RM_SMTP_USERNAME=MY_EMAIL_ADDRESS
RM_SMTP_PASSWORD=MY_PASSWORD
# Handwriting recognition
RMAPI_HWR_APPLICATIONKEY=SOME_KEY
RMAPI_HWR_HMAC=SOME_KEY
```
