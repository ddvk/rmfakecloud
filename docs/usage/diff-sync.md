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


## Deal with file lost

There was an [issue with `rmapi`](https://github.com/juruen/rmapi/issues/285)
affecting both official cloud and rmfakecloud, where files were randomly lost.

The files were not deleted on rmfakecloud side, it was possible to relink them
in the real document tree.

Two utility have been designed to help to relink:

- **`history2git15`:** it creates a git repository of the states of the tree, creating
  a commit for each generation/tree modification.
- **`relinkfile15`:** given an user and an root index name (it's sha256 hash as
  seen in `.root.history` or as commit subject of `history2git15`), it'll relink all
  files given as argument in their state at the time of the given root index.

!!! warning
    Before using those utilities, please shutdown your `rmfakecloud` instance to
    avoid concurrent modifications and possible corruptions.


### `history2git15`

Before using `history2git15`, you need to install `git` on your system.

This utility takes as argument a path to a `.root.history` file. It'll create an
`history` directory, which will contain the git repository.

```
./history2git15 /var/lib/rmfakecloud/users/ddvk/sync/.root.history
```

As it can takes a large amount of time, you can limit to the latest
modifications with the `-tail` option:

```
./history2git15 -tail 20 /var/lib/rmfakecloud/users/ddvk/sync/.root.history
```

In this example only the last 20 modifications of the tree will be saved as a
git repository.

After a successfull run, go to `/var/lib/rmfakecloud/users/ddvk/sync/history`
and use `git` to explore the differences referenced in the two files:

- **`doctree`:** this is a human readable hiereachy of the directories and
  files, along with the date of the last modification.
- **`tree`:** this is the whole tree state with all metadata accessible, in JSON
  format.


### `relinkfile15`

!!! warning
    Be sure that `rmfakecloud` is stopped before using this command. Bad things
    will happen if some devices performs synchronization while `relinkfile15`
    works.

This utility can relink in the root tree a non-deleted file.

```
DATA_DIR=/var/lib/rmfakecloud ./relink15 -user ddvk -root-hash 1c0ee6fb7fde7d09dd25b954dd9f23f950d9e25f1fbc661ca18aebf40bb14a00 "Notebook 42" "My calendar.pdf"
```

`DATA_DIR` is the same option used by `rmfakecloud`. If you don't use it with
`rmfakecloud`, make sure you are in the same directory as when you start
`rmfakecloud`: you should have a `data/` directory.

The `-user` option is the registration address used (or the name of the
directory inside `data/users`).

The `-root-hash` is the name of the file to use as the old root index.

The rest of the arguments given to the command line are the name of the file as
given by `history2git15` in `doctree` (or as `visibleName` in `tree`). This doesn't
handle directories hierarchy at the time of writing: it only matches the
filename. You need to make sure the parent directory still exists (as we said
the hierarchy is handle by metadata, not by indexes).
=======
## How does it work?

All files are saved with a sha256 name in the user `sync` directory. File can be
raw pages, document metadata, directory metadata, indexes, ...

Each file is saved along with a generation number. Each modification increment
this generation number.


### Root tree file

All files in the tree are indexed in a root index. The index references all the
index to others files and directories.

The hierarchy is given by each file metadata. The root index file doesn't handle
file hierarchy.

#### Current root file in use

It is possible to retrieve the current root index in use by consulting the
content of the file `root`. It contains the name of the file containing the root
index.

#### History and generation

Previous root indexes are kept in the directory and can be listed with the file
`.root.history`. Each line corresponds to a given generation (line 1 = generation
1, ...) and contains the date of the modification and the corresponding filename.


### File indexes

Each file, as seen on the tablet, is split in several parts (metadata, raw
pages, cached transcripted text, content, ...). File indexes store the
references of each sub-file.

Files on the tablet are referenced by an UUID. On rmfakecloud files parts are
also stores with sha256 names.
