# Welcome

rmfakecloud is a clone of the cloud sync the remarkable tablet is using, in case you want to sync/backup your files and have full control of the hosting/storage environment.

## Features

* File synchronization (compatible with revisions 1.0 and 1.5)
* Integrations with external files sources (using webdav or with a dedicated directory on local file system, instead of Google Drive and Dropbox)
* Send document by email
* Handwriting recognition
* Unlock the screen sharing feature (but it doesn't work remotely, you should use the app through USB)

It comes with a very basic web interface that let you:

* Register user
* Connect to your account
* Generate one time code for device registration
* View synchronized files
* Download PDF of the synchronized files
* Upload new documents

Please note that this project is under development and there are many features that requires to tweak configuration files directly.

## Wish List

Here is a list of tasks that still need to be accomplished:

- UI:
    * specify folder on upload
    * add/remove users
    * move files around
    * rename files
    * realtime notifications
    * document preview
    * archive / restore documents
    * share files between users
    * refactoring
    * sent emails history
- add message broker
- add db
- add blob storage
