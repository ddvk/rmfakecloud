From discord/xabean#2732 | github.com/warewolf

client: rM2, Toltec, rmfakecloud installed via opkg. Follow the device installation instructions here: https://ddvk.github.io/rmfakecloud/remarkable/setup/
server: rmfakecloud docker image, docker-compose.yml + env

# How I setup rmfakecloud for myself

General steps:

* Add a 'remarkable' user on my VPS
* I already had an existing wildcard *.mydomain.com LetsEncrypt certificate
* I `su -s /bin/bash remarkable` to become the remarkable user
* Then `mkdir rmfakecloud; cd rmfakecloud`
* Next, `mkdir data` for where rmfakecloud will store the data
* Create `docker-compose.yml` and `env` files (copy and change contents from below)
* Run `docker-compose up` to launch it in your terminal - you will need to hit Ctrl-C to stop it later
* Add apache config below to your apache's `conf.d` directory, or `sites-enabled` directory, depending on your OS
* Run `apachectl configtest` to make sure your apache config works
* Run `apachecctl graceful` to restart the webserver
* Try browsing to `https://rmfakecloud.mydomain.com` and make sure you see the rmfakecloud login page
* Add yourself a user: `docker exec rmfakecloud /rmfakecloud-docker setuser -u UserNameHere -a` -- this will print out to the screen a randomly generated password
* Log in as your username with the password it gave you, you can change it via the web UI later
* Click 'code' in the top to generate your cloud link code.
* On your rM2, go Menu -> Settings -> General -> Account -> Connect
* Enter your cloud link code from your rmfakecloud web UI
* Documents should now start syncing, you should see stuff scrolling by in your `docker-compose up` window.  Let it go for a while until it finishes.
* Make sure the cloud icon to the right of the wifi icon in the bottom left of the main "my files" screen on your rM2 tablet doesn't have an X saying the cloud is broken
* Hit Ctrl-C in your `docker-compose up` window to stop your rmfakecloud, we're done testing
* Run `docker-compose up -d` to have it run in the background like a service

Hopefully at this point, everything works?

## Docker

docker-compose.yml:
```
version: "3.4"
services:
  rmfakecloud:
    network_mode: host
    image: ddvk/rmfakecloud
    container_name: rmfakecloud
    restart: unless-stopped
    ports: 
      - 3000:3000
    env_file:
      - env
    volumes:
      - /home/remarkable/rmfakecloud/data:/data
```
  * `network_mode: host` so that the docker container can reach my mail server

env:
```
RMAPI_HWR_APPLICATIONKEY=XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX
RMAPI_HWR_HMAC=XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX
STORAGE_URL=https://rmfakecloud.mydomain.com
PORT=3000
LOGLEVEL=debug
RM_SMTP_SERVER=127.0.0.1:465
RM_SMTP_USERNAME=
RM_SMTP_PASSWORD=
RM_SMTP_FROM=ReMarkable selfhosted <rmfakecloud@mydomain.com>
JWT_SECRET_KEY=YouReallyShouldSetThisInConfigAndNotLeaveItThisStaticValueExampl
RM_SMTP_INSECURE_TLS=true
```
* JWT_SECRET_KEY - set this in config, othrewise every time your server goes up/down

## Apache

rmfakecloud.conf:
```
LoadModule proxy_wstunnel_module modules/mod_proxy_wstunnel.so
LoadModule proxy_module modules/mod_proxy.so
LoadModule proxy_http_module modules/mod_proxy_http.so

<VirtualHost rmfakecloud.mydomain.com>
  ServerName rmfakecloud.mydomain.com
  SSLEngine on
  SSLCertificateFile /etc/httpd/conf/ssl.key/mydomain.com-fullchain.pem
  SSLCertificateKeyFile /etc/httpd/conf/ssl.key/mydomain.com-cert.pem

  ProxyPass / http://localhost:3000/
  ProxyPassReverse / http://localhost:3000/
  ProxyRequests Off
  RewriteEngine on
  RewriteCond %{HTTP:Upgrade} websocket [NC]
  RewriteCond %{HTTP:Connection} upgrade [NC]
  RewriteRule ^/?(.*) "ws://localhost:3000/$1" [P,L]
</VirtualHost>                                      
```
