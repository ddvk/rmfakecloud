Since the reMarkable 2.10 release, there is a new sync method available: It
collects only differences instead of uploading the whole document. So it takes
less time to upload/download large modified documents, and it handles edition
conflicts.

!!! note
    This feature is available since rmfakecloud v0.0.5.

In order to use this feature, you'll need to update your user to activate it:

```sh
rmfakecloud setuser -u ddvk -a -s
```

The `-a` is to let/set the user admin: as the tool will remove the admin
permission if not set.

You'll then need to reconnect on your device to apply the settings, and a full
resync will automatically begin.
