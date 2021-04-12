## Without patching the binary
all needed artifacts are in `scripts/device/` folder
For Automatic script check [Automagic](scripts/device/readme.md)

- generate a CA and host certificate for `*.appspot.com`
- create the CA folder: `mkdir -p /usr/local/share/ca-certificates`
- copy the CA.crt file to `/usr/local/share/ca-certificates` and run `update-ca-certificates`
- modify the hosts file `/etc/hosts`
        - so the options are:
        1. run a reverse https proxy on the rm tablet as a service, e.g. [secure](https://github.com/yi-jiayu/secure)
            - stop xochitl `systemctl stop xochitl`
            - add to `/etc/hosts`
                ```
                127.0.0.1 hwr-production-dot-remarkable-production.appspot.com
                127.0.0.1 service-manager-production-dot-remarkable-production.appspot.com
                127.0.0.1 local.appspot.com
                127.0.0.1 my.remarkable.com
                127.0.0.1 ping.remarkable.com
                127.0.0.1 internal.cloud.remarkable.com
                ```
            - set the address of your api host:port in the reverse proxy
                `secure -cert proxy.crt -key proxy.key http://host_where_the_api_is_running:3000`
                or use the provided systemd unit file and put the config in proxycfg

            - run the host
            - run `fixsync.sh` on the device to mark all files as new (not to be deleted from the device)
            - start xochitl `systemctl start xochitl`

        2. run the fakeapi on port 443 with a certificate signed by the CA you installed
            - modify the hosts files to point to this host

        3. install only the CA certificate on the device
            - modify your DNS Server/router to resolve the aforementioned addesses to a https reverse proxy
            - install the hosts certificate on the proxy and route to the api e.g:
                - on a ubiquity router /etc/dnsmasq.d/rm.conf
                   address=/my.remarkable.com/192.168.0.10
                   etc
                - on a synology there is an application portal which you can configure as a reverse proxy
            - ***CONS*** this will affect ALL devices, but you use the mobile apps and windows clients without modifications
