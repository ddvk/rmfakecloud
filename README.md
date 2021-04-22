# rmfakecloud


rmfakecloud is fake of the cloud sync the remarkable tablet is using, in case you want to sync/backup your files and have full control of the hosting/storage environment and don't trust Google.

# Breaking Changes
- after v0.0.3 the files in `/data` will have to be manually move to the user that will be created

# Installation

## From source

Install and build the project:  
`go get -u github.com/ddvk/rmfakecloud`

run  
`~/go/bin/rmfakecloud`  


or clone an do: `go run .`  
or `make run`  
or `make` artifacts are in the `dist` folder. the Arm binaries work on pi3 / Synology etc  
or `make docker && ./rundocker.sh`  


env variables:  
`JWT_SECRET_KEY` needed for the whole auth thing to work, set something long
`PORT` port number (default: 3000)  
`DATADIR` to set data/files directory (default: data in current dir)  
`STORAGE_URL` the storage url resolvable from the device, especially if the host is behind a reverse proxy (default: http://hostname:port)  
`LOGLEVEL` default to **info** (set to **debug** for more logging or **warn**, **error** for less)

## Docker
`docker run -it --rm -p 3000:3000 ddvk/rmfakecloud` (you can pass `-h` to see the available options

# Initial Login
open `http://localhost:3000` or wherever it was installed
if no users exist, the first login creates a user

# [Tablet Setup](docs/tablet.md)
Modifications that the tablet needs

# Uploading / managing documents
For the time being, until the UI is done, [rmapi](https://github.com/juruen/rmapi) is the easist option.
```
export RMAPI_AUTH=http(s)://yourcloud
export RMAPI_DOC=http(s)://yourcloud
export RMAPI_CONFIG=~/.rmapi.fake
rmapi
```


# Resetting a user's password or creating other users
It is advisable to set the rmfakecloud's user to the user it is running under and set the sid bit (`chmod 4700 rmfakecloud`)  
also make sure the user has write permissions for the `data` directory
`DATADIR=dirwherethedatais rmfakecloud setuser -u username -p newpassword`

## Caveats
make sure to set the DATADIR env
Execute it in the context of user under witch the service is running, otherwise the profile will have the wrong user/permissions

# Handwriting Recognition
In order to get hwr running with myScript register for a developer account and set the env variables: 

`RMAPI_HWR_APPLICATIONKEY`  
`RMAPI_HWR_HMAC`

# Sending emails
Define the following env variables:

```
RM_SMTP_SERVER=smtp.gmail.com:465
RM_SMTP_USERNAME=user@domain.com
RM_SMTP_PASSWORD=plaintextpass  # Application password should work
```

If you want to provide custom FROM header for your mails, you can use:
```
RM_SMTP_FROM='"ReMarkable self-hosted" <user@domain.com>'
```

# [HTTPS HowTO](docs/https.md)


## Automagic to be run on the device
```
sh -c "$(wget https://raw.githubusercontent.com/ddvk/rmfakecloud/master/scripts/device/automagic.sh -O-)"
```

## Without patching the binary

# Caveats/ WARNING
- (applies when you don't have security) connecting to the api will delete all your files, unless you mark them as not synced `synced:false` prior to syncing (advisable just to disconnect, reconnect the cloud)
- **if you delete files from the users directory** on the host, on the next sync those will be deleted from the device
- if you delete the whole user directory (by mistake) on the host, you should disconnect the cloud from the device and reconnect it

# [TODO](docs/todo.md)
# [How the cloud works](docs/cloud.md)

# Troubleshooting
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

- if both (host and tablet) are on a wifi make sure "Client Isolation" is not actived on the AP

- check if the proxy is configured correctly
    ```
    systemctl status proxy

    #or

    journalctl -u proxy
    ```
- check if the CA cert was correctly installed
    when doing `update-ca-certificates` there should have been `1 added`
    check the logs

- check xochitls's logs, stop the service, start manually with more logging
    ```
    systemctl stop xochitl
    QT_LOGGING_RULES=xochitl.*=true xochitl | grep -A3 QUrl

    ```
    if you see *SSL Handshake failed* then something is wrong with the certs


