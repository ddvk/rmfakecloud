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
