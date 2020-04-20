# rmfakecloud


rmfakecloud is fake of the cloud sync the remarkable tablet is using. If you want to sync/backup your files and have full control of the hosting/storage environment and don't trust Google.


# Install

## From sources

Install and build the project:

`go get -u github.com/ddvk/rmfakecloud`


## Binary

TBD, some link

# Prerequisites / Device Modifications

## Without patching the binary

Install a root CA on the device, you can use the ones inlcuded in this repo, but it's better you could generate your own
- generate a CA and host certificate for *.appspot.com []()
- copy the CA.crt file to `/usr/local/share/ca-certificates` and run `update-ca-certificates`
- modify the hosts file `/etc/hosts`
	- so the options are:
        1. run a reverse https proxy on the device as a service, i.g. [secure](github.com/yi-jiayu/secure)
            - add to /etc/hosts
                ```
                127.0.0.1 service-manager-production-dot-remarkable-production.appspot.com
                127.0.0.1 local.appspot.com
                127.0.0.1 my.remarkable.com
                ```
            - set the address of your host:port in the reverse proxy
                `secure -cert example.org.bundle.crt -key example.org.key http://10.11.99.4:3000`
            - run the host
		2. run the fakeapi on port 443 with a certificate signed by the CA you installed and resolve 
		3. run a reverse proxy on the host and route to the api


## Patching the binary
- some script to set an address

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
