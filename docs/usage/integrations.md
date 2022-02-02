Integration is the [feature added in reMarkable
2.10](https://support.remarkable.com/hc/en-us/articles/4406214540945)
that allows to browse, download and upload document from location
outside of the tablet.

!!! warning
    Integrations are still work in progress, no UI exists yet.

## WebDAV

It can be used with any WebDAV services, for example a Nextcloud/Owncloud instance.

Add this to your [`.userprofile`](userprofile.md):

```yaml
integrations:
  - provider: webdav
    id: [generate some uuid]
    name: [some name]
    username: [username]
    password: [password]
    address: [webdavaddrss]
    insecure: [true/false] (to skip certificate checks)
```


For example:

```yaml
integrations:
  - provider: webdav
    id: fLAME8YBm5uFJ89GKRAFkGjk7hJw0heow045kfhc
    name: Home Nextcloud
    address: https://home.example.com/remote.php/dav/files/user42/
    username: user42
    password: password4242
```



## Local File System

!!! warning
    Experimental and not suited for multiple users yet

You can share a dedicated path on your system. This can be a simple directory or a mount point using FUSE or whatever.

Add this to your [`.userprofile`](userprofile.md):

```yaml
integrations:
  - provider: localfs
    id: [generate some uuid]
    name: [some name]
    path: /some/path/with/files
```
