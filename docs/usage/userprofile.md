Have a look inside `data` directory ([`DATADIR`](../install/configuration.md)):
you'll find under `data/users/` a directory by user (since v0.0.3). The
directory name is expected to be the username given in the webUI login form.


## User Settings

rmfakecloud stores user configuration (password, email, options, ...) in a file
inside its directory, named `.userprofile`. This is a hidden file.

This file, written in YAML, have the following relevant entries:

| Entry      | Description |
|------------|-------------|
| `password` | Password to access the account (in Argon2 format) |
| `name` | Name displayed in the webui |
| `isadmin` | Boolean indicating if the user can perform administration tasks (currently managing user accounts) |
| `sync15` | Boolean value that indicates if the user is using the [diff synchronization](diff-sync.md) (aka. sync 1.5) |
| `integrations` | Array with the user integrations. See [Integrations](integrations.md) |


### Edit settings through CLI

Use the same binary as for launching the server: it takes some specials commands described bellow.

When using the Docker image, you can run :

```sh
docker exec rmfakecloud /rmfakecloud-docker special-command
```

#### `rmfakecloud listuser`

This commands lists existing users.

#### `rmfakecloud setuser`

This commands edit or create account.

To create/update an admin account `ddvk`:

```sh
rmfakecloud setuser -u ddvk -a
```

To reset a password:

```sh
read -s -p "New password: " NEWPASSWD && rmfakecloud setuser -u ddvk -p "${NEWPASSWD}"
```


## Directory Structure

In a user directory, there are files like `[UUID].metadata` and `[UUID].zip`
(if you are not using [sync 1.5](diff-sync.md)): this corresponds to your raw
documents on your tablet.

There is also a `trash` directory, containing deleted files on the tablet, in
its trash.

If you are using [sync 1.5](diff-sync.md), the magic happen in the `sync`
directory.
