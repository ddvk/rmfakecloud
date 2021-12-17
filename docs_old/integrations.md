# WebDav (nextCloud)
This is still work in progress and no ui exists yet

## Add this to your .userprofile
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


# Local File System
Experimental and not suited for multiple users yet
## Add this to your .userprofile
```yaml
integrations:
  - provider: localfs
    id: [generate some uuid]
    name: [some name]
    path: /some/path/with/files
```
