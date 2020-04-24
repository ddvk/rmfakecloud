# rmfakecloud


rmfakecloud is fake of the cloud sync the remarkable tablet is using. If you want to sync/backup your files and have full control of the hosting/storage environment and don't trust Google.

# Status 
early prototype (sync and notifications work). no security and a lot of quick and dirty hacks.

# Install

## From sources

Install and build the project:

`go get -u github.com/ddvk/rmfakecloud`

run
`~/go/bin/rmfakecloud`

or clone an do: `go run .`

env variables
```
PORT to change the port number (default: 3000)
DATA_DIR to set data/files directory (default: data in current dir)
```


# Prerequisites / Device Modifications

## Without patching the binary
all needed artifacts are in `device/` folder

Install a root CA on the device, you can use the ones inlcuded in this repo, but it's better you could generate your own
- generate a CA and host certificate for *.appspot.com []()
- create the CA folder: `mkdir -p /usr/local/share/ca-certificates`
- copy the CA.crt file to `/usr/local/share/ca-certificates` and run `update-ca-certificates`
- modify the hosts file `/etc/hosts`
	- so the options are:
        1. run a reverse https proxy on the rm tablet as a service, e.g. [secure](https://github.com/yi-jiayu/secure)
            - stop xochitl `systemctl stop xochitl`
            - add to /etc/hosts
                ```
                127.0.0.1 service-manager-production-dot-remarkable-production.appspot.com
                127.0.0.1 local.appspot.com
                127.0.0.1 my.remarkable.com
                ```
            - set the address of your api host:port in the reverse proxy
                `secure -cert example.org.bundle.crt -key example.org.key http://10.11.99.4:3000`
                or use the provided systemd unit file and put the config in proxycfg

            - run the host
            - run `fixsync.sh` on the device to mark all files as new (not to be deleted from the device)
            - start xochitl `systemctl start xochitl`
		2. run the fakeapi on port 443 with a certificate signed by the CA you installed and resolve 
        - modify the hosts files to point to this host
        3. install only the CA certificate on the device
        - modify your DNS Server/router to resolve the aforementioned addesses to a https reverse proxy
        - install the hosts certificate on the proxy and route to the api e.g:
            - on a ubiquity router /etc/dnsmasq.d/rm.conf
               address=/my.remarkable.com/192.168.0.10
               etc
            - on a synology there is an application portal which you can configure as a reverse proxy
        - ***CONS*** this will affect ALL devices, but you use the mobile apps and windows clients without modifications


# Caveats/ WARNING
- connecting to the api will delete all you files, unless you mark them as not synced `synced:false` prior to syncing

# TODO

- [ ] auth / authz
- [ ] multi tenant
- [ ] fix go stuff
- [ ] storage provider
- [ ] db
- [ ] liveview
- [ ] refactor


# How the cloud sync works or (will I lose my files) requested by @torwag
(my interpretation and flawed observations)

Given a new unregistered device, all the files that are generated locally have a status `synced:false`

Registration:

the device sends a post request to: `my.remarkable.com/token/json/2/device/new`
containing the random key (currently any key will be accepted) and gets a device token

with the device token it obtains expiring access tokens: `my.remarkable.com/token/json/2/user/new`

having a user access token: 

sends a request to the services locator to get the urls of additional services:
`/service/json/1/(web|mail|notification|storage)`(now always local.appspot.com)
it gets a list of all documents


Gets the list of documents from the server: `/document-storage/json/2/docs`
the order may be not correct at all:
- ***deletes*** all documents not in the list and having `synced: false`
- ***deletes*** all documents from the cloud that have `deleted: true`
- applies renames, page changes
- sends all new and marks them `synced: true`
- sends all modified documents (having a greater Version number?)
- it doesn't re-download any changed/newer documents


So if you just point the device to a new empty server, all documents will be deleted from the device. 
Going back will again, delete all documents and put what was on the server

