# Desktop Client

This page covers using the official reMarkable desktop apps with a self-hosted rmfakecloud.

There are two approaches:

1. Manual setup (edit the hosts file).
2. Patch the desktop app (`RMHook` or `RMHook-Win`).

## Manual Setup (Hosts File)

This method keeps the official desktop app but redirects traffic by editing the hosts file. It is useful if you do not want to patch the app.

### macOS /etc/hosts

1. Open `/etc/hosts` as root.
2. Point the reMarkable cloud domains to your rmfakecloud IP.

Example (replace `203.0.113.5` with your rmfakecloud public IP):

```
203.0.113.5 hwr-production-dot-remarkable-production.appspot.com
203.0.113.5 service-manager-production-dot-remarkable-production.appspot.com
203.0.113.5 local.appspot.com
203.0.113.5 my.remarkable.com
203.0.113.5 ping.remarkable.com
203.0.113.5 internal.cloud.remarkable.com
203.0.113.5 webapp-prod.cloud.remarkable.engineering
203.0.113.5 eu.tectonic.remarkable.com
203.0.113.5 backtrace-proxy.cloud.remarkable.engineering
```

This list follows the guidance in the "Getting Traffic To The Server" section from:
https://blog.scottlabs.io/2025/10/selfhosting-a-remarkable-cloud/#getting-traffic-to-the-server

### Windows hosts file

1. Open `C:\Windows\System32\drivers\etc\hosts` in an elevated editor (Run as Administrator).
2. Add the same domain entries as above, pointing to your rmfakecloud IP.

### Linux /etc/hosts

1. Open `/etc/hosts` as root.
2. Add the same domain entries as above, pointing to your rmfakecloud IP.

### Certificates

- Add the CA certificate to macOS Keychain and `server.crt` / `server.key` installed on rmfakecloud.

## macOS Desktop via RMHook

If you only need to point the official macOS desktop app at your self-hosted instance, you can patch it with [RMHook](https://github.com/NohamR/RMHook). The tool injects a custom dylib into the reMarkable Desktop bundle so all cloud traffic is redirected to your rmfakecloud host without editing `/etc/hosts` and adding CA certificates.

There is an auto script provided on the repo as well as a manual setup.

### Auto Installation

Run in a terminal:

```
bash <(curl -sL https://raw.githubusercontent.com/NohamR/RMHook/refs/heads/main/scripts/auto-install.sh)
```

## Windows / Linux via RMHook-Win

[RMHook-Win](https://github.com/NohamR/RMHook-Win) is a Windows port of RMHook for the reMarkable Desktop application. It builds a proxy DLL that hooks Qt network APIs and redirects reMarkable cloud traffic to a self-hosted rmfakecloud server.

RMHook-Win intercepts the reMarkable Desktop app's Qt networking layer and patches outgoing requests to the configured host and port. It is designed for the Windows reMarkable Desktop client and uses a DLL proxy for `paho-mqtt3as.dll`.

### Auto Installation

Run in a PowerShell terminal with administrator privileges:

```
irm https://raw.githubusercontent.com/NohamR/RMHook-Win/refs/heads/main/scripts/download-and-install.ps1 | iex
```