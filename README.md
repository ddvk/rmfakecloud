# rmfakecloud
rmfakecloud is a clone of the cloud sync the remarkable tablet is using, in case you want to sync/backup your files and have full control of the hosting/storage environment.

## Breaking Changes
- after v0.0.3 the files in `/data` will have to be manually moved to the user that will be created
- with v0.0.5 the new diff sync15 is added as an option, in order to use it modify the user with `setuser -u user -s`  
  or modify the profile and add `sync15:true`  
  a full resync will be needed (the tablet will do it), the old files are kept as they were and everying is put in a new directory

## Installation

### From source

Install and build the project, under Linux:  
`git clone https://github.com/ddvk/rmfakecloud`  
`make`

`make` will just build the x64 binary. If you need binaries for other architectures, you should instead run `make all`.

run  
`~/dist/rmfakecloud-x64`

or clone an do: `go run ./cmd/rmfakecloud`  
or `make run`  
or `make all` artifacts are in the `dist` folder. the Arm binaries work on pi3 / Synology etc  
or `make docker && ./rundocker.sh`


### From Docker
`docker run -it --rm -p 3000:3000 -e JWT_SECRET_KEY='something' ddvk/rmfakecloud` (you can pass `-h` to see the available options

### Configuration Environment Variables
`JWT_SECRET_KEY` needed for the whole auth thing to work, set something long  
`STORAGE_URL` controls whether file upload/download goes through the local proxy or directly. the address of rmfakecloud **as visible from the tablet**, especially if the host is behind a reverse proxy or in a container (default: http://hostname:port)  
`PORT` port number (default: 3000)  
`DATADIR` to set data/files directory (default: data in current dir)  
`LOGLEVEL` default to **info** (set to **debug** for more logging or **warn**, **error** for less)
**`RM_HTTPS_COOKIE=1`** UI, send auth cookies only via https 

## Initial Login
open `http://localhost:3000` or wherever it was installed
if no users exist, the first login creates a user

## [Tablet Setup](docs/tablet.md)
Modifications that the tablet needs


## [Integrations](docs/integrations.md)
3rd party integrations

## Uploading / managing documents
The UI is still wip, for cli [rmapi](https://github.com/juruen/rmapi) is quite good.
```
export RMAPI_AUTH=http(s)://yourcloud
export RMAPI_DOC=http(s)://yourcloud
export RMAPI_HOST=http(s)://yourcloud #the only one needed after 0.0.16
export RMAPI_CONFIG=~/.rmapi.fake
rmapi
```

## Resetting a user's password or creating other users
It is advisable to set the rmfakecloud's user to the user it is running under and set the sid bit (`chmod 4700 rmfakecloud`)  
also make sure the user has write permissions for the `data` directory
`DATADIR=dirwherethedatais rmfakecloud setuser -u username -p newpassword`

### Caveats
make sure to set the `DATADIR` env
Execute it in the context of user under witch the service is running, otherwise the profile will have the wrong user/permissions

## Handwriting Recognition
In order to get hwr running with myScript register for a developer account and set the env variables: 

`RMAPI_HWR_APPLICATIONKEY`  
`RMAPI_HWR_HMAC`

## Sending emails
Set the following env variables:

```
RM_SMTP_SERVER=smtp.gmail.com:465
RM_SMTP_USERNAME=user@domain.com
RM_SMTP_PASSWORD=plaintextpass  # Application password should work
```

If you want to provide custom FROM header for your mails, you can use:
```
RM_SMTP_FROM='"ReMarkable self-hosted" <user@domain.com>'
```


## Development
run `./dev.sh` which will start the UI and backend

## [HTTPS HowTO](docs/https.md)

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

## TODO
- [ ] UI specify folder on upload
- [ ] UI add/remove users
- [ ] UI move files around
- [ ] UI rename files
- [ ] UI realtime notifications
- [ ] UI document preview
- [ ] UI archive / restore documents
- [ ] UI share files between users
- [ ] UI refactoring
- [ ] UI sent emails history
- [ ] add message broker
- [ ] add db
- [ ] add blob storage
