# rmfakecloud
This is a replacement of the cloud, in case you want to sync/backup your files and have full control of the hosting environment.

## [Docs](https://ddvk.github.io/rmfakecloud/)

## NB
For Tablet SW > 3.X, rendering of the notebooks is not yet supported.

## Breaking Changes
- after v0.0.3 the files in `/data` will have to be manually moved to the user that will be created
- with v0.0.5 the new diff sync15 is added as an option, in order to use it modify the user with `setuser -u user -s`  
  or modify the profile and add `sync15:true`  
  a full resync will be needed (the tablet will do it), the old files are kept as they were and everying is put in a new directory

## Development
run `./dev.sh` which should start the UI and backend

### Caveats/ WARNING
- (applies when you don't have security, version <= 0.0.3) connecting to the api will delete all your files, unless you mark them as not synced `synced:false` prior to syncing (advisable just to disconnect, reconnect the cloud)
- **if you delete files from the users directory** on the host, on the next sync those will be deleted from the device
- if you delete the whole user directory (by mistake) on the host, you should disconnect the cloud from the device and reconnect it
- after an official update, the proxy and hosts file changes will be removed, the tablet will automatically disconnect from the cloud (by sending an invalid token to the official cloud and getting 403)
  just reinstall the proxy and reconnect to your cloud

## Troubleshooting
- check the connectivity between the tablet and the host:
    ping my.remarkable.com (should be localhost)
    ping local.remarkable.com (should be localhost)
    ping thehostpc
    wget -qO- http://host:3000 (or relevant ports, should get Working...)
    wget -qO- https://local.appspot.com (should get Working...)
    
- check that the proxy is running and certs are installed:
    ```
    echo Q | openssl s_client -connect localhost:443  -verify_hostname local.appspot.com -CAfile /etc/ssl/certs/ca-certificates.crt 2>&1 | grep Verify
    ```
    You should see: *Verify return code: 0 (ok)*

- if both (host and tablet) are on a wifi make sure "Client Isolation" is not activated on the AP

- check if the proxy is configured correctly
    ```
    systemctl status proxy

    #or

    journalctl -u proxy
    ```
- check whether the CA cert was installed correctly
    when doing `update-ca-certificates` there should have been `1 added`
    check the logs

- check xochitls's logs, stop the service, start manually with more logging
    ```
    systemctl stop xochitl
    QT_LOGGING_RULES=rm.network.*=true xochitl | grep -A3 QUrl

    ```
    if you see *SSL Handshake failed* then something is wrong with the certs

