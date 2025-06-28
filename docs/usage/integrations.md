Integration is the [feature added in reMarkable
2.10](https://support.remarkable.com/hc/en-us/articles/4406214540945)
that allows to browse, download and upload document from location
outside of the tablet.

You can edit your integrations using the Integration tab in the UI.


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



## Messaging webhook

Messaging are a type of integration added in software 3.17.

Originally designed for Slack, this feature allows you to send your current sheet as an attachment to a Slack Canvas. The default behavior uses AI to transcribe the handwritten content on your sheet and posts both the text and the image to Slack.

The Webhook integration extends this capability by sending your sheet to an external automation platform (like [n8n](https://n8n.io/), [Make.com](https://www.make.com/), ...) or custom service. This is especially useful if you want to: use your own AI pipeline or don't want AI to be involved at all, or store and process sheets in a custom backend, ...

The webhook gives you full control: you decide what happens with your data.
