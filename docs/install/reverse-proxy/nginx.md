From @zeigerpuppy

> edit: had to turn off proxy buffering and add `proxy_redirect http:// https://;` for the assets to load properly

I have rmfakecloud up and running (reMarkable 2 client, Debian 9 server).  It's working great, sync, emails and handwriting recognition are all good.
I am using the local proxy config and have now tested a working HTTPS connection for increased security (comments appreciated).
Currently, my understanding is that in the default config, the proxy is just establishing an HTTP proxy connection as the rmfakecloud is served on http://server:3000.
I would like to have this working on public IP networks too and have set up a NAT rule to forward port 3000 to my local server.  This works but I guess it's all unencrypted.

> note that once HTTPS is working, direct forwarding of port 3000 should be disabled!

So, to get it working via HTTPS, I think all we need to do is to set up a reverse HTTPS proxy on the server.
NB: **I initially tried this with Apache2 but couldn't get the websockets working**.  The error on the server was this:

```
INFO[0387] accepting websocket abc
INFO[0387] upgrade:websocket: the client is not using the websocket protocol: 'upgrade' token not found in 'Connection' header
INFO[0387] closing the ws
INFO[0387] [GIN] 2020/11/17 - 13:39:41 | 400 |      201.78µs |       127.0.0.1 | GET      "/notifications/ws/json/1"
```

So, I tried with an nginx reverse proxy with the following config:

```nginx
server {
    # increase max request size (for large PDFs)
    client_max_body_size 200M;
    server_name rmfakecloud.server.net;
    client_body_buffer_size 1280K;
    proxy_buffering off;
    proxy_request_buffering off;

    listen 443 ssl; # managed by Certbot
    ssl_certificate /etc/letsencrypt/live/rmfakecloud.server.net/fullchain.pem; # managed by Certbot
    ssl_certificate_key /etc/letsencrypt/live/rmfakecloud.server.net/privkey.pem; # managed by Certbot
    include /etc/letsencrypt/options-ssl-nginx.conf; # managed by Certbot
    ssl_dhparam /etc/letsencrypt/ssl-dhparams.pem; # managed by Certbot

    location / {
        proxy_pass http://localhost:3000/;
        proxy_set_header Host $host;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_redirect http:// https://;

        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
  }

  upstream ws-backend {
    # enable sticky session based on IP
    ip_hash;
    server localhost:3000;
}

```

That seems to work well.
There were two other config changes to make:

1. set the STORAGE_URL on the server: `export STORAGE_URL=https://rmfakecloud.server.net`
2. change the proxy URL on the device: 

stop services
```
systemctl stop xochitl
systemctl stop proxy 
```

edit the proxy address
```
nano /etc/systemd/system/proxy.service
```

change the line `ExecStart` to have the new address
```
...
ExecStart=/home/root/scripts/rmfakecloud/secure -cert /home/root/scripts/rmfakecloud/proxy.crt -key /home/root/scripts/rmfakecloud/proxy.key https://rmfakecloud.server.net
...
``` 

reload and start services

```
systemctl daemon-reload
systemctl start proxy
systemctl start xochitl
```

I think this is all good, happy to hear feedback but I think we should amend a section on the README to show how to configure with HTTPS.

Now, the only thing needed is starting the server automatically....
