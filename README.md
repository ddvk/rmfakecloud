# rmfakecloud

This is a replacement of the cloud, in case you want to sync/backup your files and have full control of the hosting environment.

See the [project documentation](https://ddvk.github.io/rmfakecloud/) for setup and configuration.

## Supported Devices

| Device               | Is Supported |
| -------------------- | ------------ |
| reMarkable 1         | ✅           |
| reMarkable 2         | ✅           |
| reMarkable Paper Pro | ✅           |
| reMarkable Paper Pro Move | ✅           |

The current release of rmfakecloud supports file synchronization up to **reMarkable software 3.22.0**. Newer releases have not been tested yet.

Use the `rmfakecloud-proxy` from [toltec](https://github.com/toltec-dev/toltec/). [More in the doc](https://ddvk.github.io/rmfakecloud/remarkable/setup/).


## Feature Parity With Official Cloud

| Features | Supported | Notes |
| -------- | --------- | ----- |
| File synchronization (1.0) | ✅ |  |
| File synchronization (1.5, 2, 3, 4) | ✅ |  |
| [Send document by email](https://ddvk.github.io/rmfakecloud/install/configuration/#email-settings) | ✅ |  |
| [Handwriting recognition](https://ddvk.github.io/rmfakecloud/install/configuration/#handwriting-recognition) | ✅ |  |
| Handwriting search | ❌ |  |
| Screen sharing | 🟡 | unlocked on reMarkable but doesn't work remotely |
| [Storage integrations](https://ddvk.github.io/rmfakecloud/usage/integrations/) | ✅ |  |
| Integration with Dropbox | 🟡 | [WIP](https://github.com/ddvk/rmfakecloud/blob/master/internal/integrations/dropbox.go) |
| Integration with Google Drive | 🟡 | [WIP](https://github.com/ddvk/rmfakecloud/pull/241) |
| Integration with OneDrive | ❌ |  |
| Integration with WebDAV | ✅ | Nextcloud, Owncloud, ... |
| Integration with FTP | ✅ |  |
| Messaging integrations | ✅ |  |
| [Messaging integration through webhook](https://ddvk.github.io/rmfakecloud/usage/integrations/#messaging-webhook) | ✅ |  |
| Messaging integration to Slack | 🟡 | Not directly, use a webhook with zapier/make/n8n |
| Archive document to cloud | 🟡 | It works but the information is not saved |
| Document rendering in web interface | ✅ |  |
| v6 file format support (software 3.0+) | ✅ | Native in-process rendering with rmc-go |


## Breaking Changes

- For SW >= 3.15 `STORAGE_URL` should not be set (or only https://some.ho.st without a port should be used)
- after v0.0.3 the files in `/data` will have to be manually moved to the user that will be created
- with v0.0.5 the new diff sync15 is added as an option, in order to use it modify the user with `setuser -u user -s`
  or modify the profile and add `sync15:true`
  a full resync will be needed (the tablet will do it), the old files are kept as they were and everything is put in a new directory

## v6 File Format Support

rmfakecloud natively supports v6 file format (introduced in reMarkable software 3.0+) for PDF export via the web UI.

### How It Works

- **v5 files** (software < 3.0): Rendered using built-in rmapi library
- **v6 files** (software >= 3.0): Rendered natively using integrated rmc-go library with Cairo

### Requirements

**Docker (recommended)**: All dependencies included automatically.

**Manual installation**:
- Cairo development libraries for building:
  - macOS: `brew install cairo pkg-config`
  - Ubuntu/Debian: `apt-get install libcairo2-dev pkg-config`
  - Fedora: `dnf install cairo-devel`
- Cairo runtime library for running (installed automatically on most systems)

### Building

```bash
# Build with native v6 support (Cairo)
make build-cairo

# Or for standard build
make build
```

### Docker Usage

The Docker image includes native v6 support:

```bash
docker run -d -p 3000:3000 \
  -v $PWD/data:/data \
  ddvk/rmfakecloud:latest
```

### Performance

- **v6 rendering**: ~100-500ms per page (in-process, no external dependencies)
- **v5 rendering**: Similar performance with rmapi
- **Caching**: Results cached for faster subsequent access

### Known Limitations

- Multi-page v6 documents: Fully supported
- Text rendering: Fully supported

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
- check sync logs
   ```
   journalctl -u rm-sync
   ```
