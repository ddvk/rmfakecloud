# Device Setup

There are several ways to make it work, choose whatever works for you

!!! important
    Whenever the tablet receives a system update, the cloud
    connection will break, and [will have to be
    reenabled](#reenabling-after-a-system-update).

## Automatic

### toltec

Install using [toltec](https://toltec-dev.org/).

```commandline
opkg install rmfakecloud-proxy
rmfakecloudctl set-upstream <URL>
rmfakecloudctl enable
```

The `<URL>` above has to be the same as in the `STORAGE_URL` server configuration.

### rmfakecloud-proxy script

Get the installer from: [rmfakecloud-proxy](https://github.com/ddvk/rmfakecloud-proxy/releases)
scp to the device and run it

**reMarkable 1 & reMarkable 2**
```shell
scp installer-rm12.sh root@10.11.99.1:
ssh root@10.11.99.1
chmod +x ./installer-rm12.sh
./installer-rm12.sh install
```

reMarkable Paper Pro & Paper Pro Move
```shell
scp installer-rmpro.sh root@10.11.99.1:
ssh root@10.11.99.1
chmod +x ./installer-rmpro.sh
./installer-rmpro.sh install
```

the automagic script doesn't work anymore, because of the busybox wget tls

```commandline
sh -c "$(wget https://raw.githubusercontent.com/ddvk/rmfakecloud/master/scripts/device/automagic.sh -O-)"
```

## Manual

### Installing a proxy on devices

A reverse proxy [rmfakecloud-proxy](https://github.com/ddvk/rmfakecloud-proxy/releases) has to be installed
run rmfakecloud on whichever port you want, you can use either HTTP (not recommended) or HTTPS, generate a new cert for the url you chose e.g with Let's Encrypt

Steps (done by the automagic scripts):

- generate a CA and host certificate for `*.appspot.com`
- create the CA folder: `mkdir -p /usr/local/share/ca-certificates`
- copy the CA.crt file to `/usr/local/share/ca-certificates` and run `update-ca-certificates`
- modify the hosts file `/etc/hosts`
- Run a reverse https proxy on the rm tablet as a service, e.g. [secure](https://github.com/yi-jiayu/secure),
- stop xochitl `systemctl stop xochitl`
- add the following entries to `/etc/hosts`

```
127.0.0.1 hwr-production-dot-remarkable-production.appspot.com
127.0.0.1 service-manager-production-dot-remarkable-production.appspot.com
127.0.0.1 local.appspot.com
127.0.0.1 my.remarkable.com
127.0.0.1 ping.remarkable.com
127.0.0.1 internal.cloud.remarkable.com
127.0.0.1 eu.tectonic.remarkable.com
127.0.0.1 backtrace-proxy.cloud.remarkable.engineering
127.0.0.1 dev.ping.remarkable.com
127.0.0.1 dev.tectonic.remarkable.com
127.0.0.1 dev.internal.cloud.remarkable.com
127.0.0.1 eu.internal.tctn.cloud.remarkable.com
```

- set the address of your api host:port in the reverse proxy
    `secure -cert proxy.crt -key proxy.key http(s)://host_where_the_api_is_running:someport`
    or use the provided systemd unit file and put the config in proxycfg
- set the `STORAGE_URL` to point to this address (this thing the device can resolve/see e.g the reverse proxy, public dns etc)
- run the host
- run `fixsync.sh` on the device to mark all files as new (not to be deleted from the device)
- start xochitl `systemctl start xochitl`

Windows/Mac Desktop Client:

- modify the hosts file (`\system32\drivers\etc\hosts`) add the same entries as on the tablet
- run a reverse proxy on the host or somewhere else pointing it to rmfakecloud with the same certs
- profit

**PROS**: easy setup, you can use whichever port you want, you can get a real trusted ca cert from let's encrypt, if running in a trusted network you may chose to use HTTP  
**CONS**: you have to configure HTTPS on the host yourself, additional Desktop config

### Modify device /etc/hosts

Connect to the host directly, without a reverse proxy, with HTTPS on :443

Steps:

- generate the certs from Variant 1, you get them (proxy.crt, proxy.key, ca.crt) and trust the ca.crt
- run rmfakecloud with:

```
TLS_KEY=proxy.key
TLS_CERT=proxy.crt
STORAGE_URL=https://local.appspot.com
```

- modify `/etc/hosts` but use the rmfakecloud's ip instead of 127.0.0.1

Windows/Mac Desktop Client:

- trust the `ca.crt`  (add it to Trusted Root CA, use cert.msc)
- modify the hosts file (`\system32\drivers\etc\hosts`) add the same entries as on the tablet
- profit

**PROS**: you can use the Windows/Mac clients, no need for a proxy on the device  
**CONS**: a bit harder to setup, each host has to trust the ca and modify the hosts file, you have to use port 443

### Edit router DNS entries

Same as [the previous method](#modify-device-etchosts), but instead of modifying any hosts file, make the changes on your DNS/router:

- add the host entries directly on your router (Hosts in OpenWRT)
- trust the ca.crt
- profit

**PROS**: a bit easier, you can you even the mobile apps if you manage to install the root ca  
**CONS**: you can't use the official cloud anymore due to the mangled DNS

## Reenabling after a system update

After a system update, you'll need to go through the steps to setup your device again, as any changes you made to it's configuration will have been lost.

# Login

After you installed the proxy, you will need to login to your account on your device.

1. Click `Menu > General > Account`.
2. Click on `Setup Account`.
3. On your main computer:
    1. Login to the rmfakecloud Web UI (if no proxy used, the same as the `STORAGE_URL` value in the server configuration).
    2. Press the `Code` link in the menu.
    3. Press the `Generate Code` button.
4. Enter the shown code on your device.
5. To check that sync is working correctly. Go to `Menu > Storage` and press `Check Sync`.
