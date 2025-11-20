The configuration is made through environment variables.

## General configuration

| Variable name     | Description |
|-------------------|-------------|
| `JWT_SECRET_KEY`  | The secret key used to sign the authentication token.<br>If you don't provide it, a random secret is generated, invalidating all connections established previously to be closed.<br>A good secret is for example: `openssl rand -base64 48` |
| `STORAGE_URL`     | It controls whether file upload/download goes through the local proxy or to an external server. It's the full address (protocol, host, port, path) of rmfakecloud **as visible from the tablet**, especially if the host is behind a reverse proxy or in a container. Example: `http://192.168.2.3:3000` (default: `https://local.appspot.com`), on SW 3.15 only https without port will work |
| `PORT`            | listening port number (default: 3000) |
| `DATADIR`         | Set data/files directory (default: `data/` in current dir) |
| `LOGLEVEL`        | Set the log verbosity. Default is **info**, set to **debug** for more logging or **warn**, **error** for less |
| `RM_HTTPS_COOKIE` | For the UI, force cookies to be available only via https |
| `RM_TRUST_PROXY`  | Trust the proxy for client ip addresses (X-Forwarded-For/X-Real-IP) default false |
| `HASH_SCHEMA_VERSION` | Hash tree schema version: "3" or "4" (default: 4) |

## Handwriting recognition

To use the handwriting recognition feature, you need first to create a free account on <https://developer.myscript.com/> (up to 2000 free recognitions per month).

Then you'll obtains an application key and its corresponding HMAC to give to rmfakecloud:

| Variable name              | Description |
|----------------------------|-------------|
| `RMAPI_HWR_APPLICATIONKEY` | Application key obtained from myscript |
| `RMAPI_HWR_HMAC`           | HMAC obtained from myscript |
| `RMAPI_HWR_LANG_OVERRIDE`  | Optional: Use this if you want your handwriting to be recognized as a different language. This variable accepts a locale code (e.g., zh_CN). Refer to [this page](https://app-support.myscript.com/support/solutions/articles/16000086001-supported-languages) for supported languages.|

## Email settings

To be able to send email from your reMarkable, fill the following variables:

| Variable name          | Description |
|------------------------|-------------|
| `RM_SMTP_SERVER`       | The SMTP server address in  host:port format |
| `RM_SMTP_USERNAME`     | The username/email for login |
| `RM_SMTP_PASSWORD`     | Plaintext password (application password should work) |
| `RM_SMTP_FROM`         | Custom `From:` header for the mails (eg. `ReMarkable self-hosted <remarkable@my.example.net>`). If this override is set, the user's email address is instead put as `Reply-To` |
| `RM_SMTP_HELO`         | Custom HELO, if your email provider needs it |
| `RM_SMTP_NOTLS` | don't use tls |
| `RM_SMTP_STARTTLS` | use starttls command, should be combined with NOTLS. in most cases port 587 should be used |
| `RM_SMTP_INSECURE_TLS` | If set, don't check the server certificate (not recommended) |

## Screen sharing

| Variable name     | Description |
|-------------------|-------------|
| `MQTT_PORT`       | Port for MQTT broker (default: 8883) |
| `ICE_SERVERS`     | JSON array of WebRTC ICE servers. Default: none. Format: `[{"urls":["stun:stun.l.google.com:19302"]}]` or with TURN: `[{"urls":["turn:turn.example.com:3478"],"username":"user","credential":"pass"}]` |
| `TLS_CERT`          | `path/to/cert`, required for screen sharing |
| `TLS_KEY`           | `/path/to/key`, required for screen sharing |

TLS certificates are required for screen sharing. Desktop apps may not use system certificate store for MQTT.  
Requires overriding DNS for `vernemq-prod.cloud.remarkable.engineering` to point to your rmfakecloud instance and using a TCP (not HTTP) reverse proxy.  
Without `ICE_SERVERS` set, screen sharing will work over USB and if the tablet and desktop app are on the same network.

### Reverse proxy for MQTT (Screen sharing)

MQTT uses TCP with TLS. Typical reverse proxies require TCP stream forwarding rather than HTTP proxying.

#### nginx (stream module)

```nginx
stream {
    upstream mqtt {
        server rmfakecloud:8883;
    }

    server {
        listen 8883;
        proxy_pass mqtt;
        proxy_connect_timeout 5s;
    }
}
```

#### Traefik (TCP router)

```yaml
tcp:
  routers:
    mqtt:
      rule: "HostSNI(`*`)"
      service: mqtt
      entryPoints:
        - mqtt
  services:
    mqtt:
      loadBalancer:
        servers:
          - address: "rmfakecloud:8883"

entryPoints:
  mqtt:
    address: ":8883"
```
