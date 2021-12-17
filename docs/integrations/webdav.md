# WebDAV (nextcloud)

This is still work in progress and no ui exists yet.
It can used with webdav services, for example provided by a self hosted nextcloud instance.

Add this to your .userprofile:
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
