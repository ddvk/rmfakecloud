# Fail2ban

For a public server [fail2ban](https://www.fail2ban.org/wiki/index.php/Main_Page) adds some security by banning ip's after few (configurable) failed login attempts.
Assuming rmfakecloud is running in docker via systemd and logs to the syslog (journalctl) and fail2ban is already installed and setup.
Instructions install and setup fail2ban in the documentation of the used operating system or at https://github.com/fail2ban/fail2ban#installation .
rmfakecloud needs to trust the reverse proxy in use, i.e. add `RM_TRUST_PROXY=1` to the docker environment,
see [configuration](configuration.md).

## Jail
First it is necessary to define a jail, e.g. in `/etc/fail2ban/jail.local`:
```
[rmfakecloud]
enabled   = true
filter   = rmfakecloud
action   = iptables-multiport[name=HTTP, port="http,https"]
journalmatch = _SYSTEMD_UNIT=docker-rmfakecloud.service
```
where it is necessary to change "docker-rmfakecloud.service" to the name of the systemd service used to run the server
and the ban action appropriate to the setup of the host.

## Filter
Additionally the filter
```
[Definition]
failregex = ^.*, login failed ip:\s+<ADDR>.*$
```
in `/etc/fail2ban/filter.d/rmfakecloud.conf` tells fail2ban which lines are relevant.

After creating the necessary configuration, restarting fail2ban loads the changes.
