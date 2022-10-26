#!/bin/bash
# rudimentary smoke test
set -e

rm -fr ../data/users/test/sync/*
rm -fr ../data/users/test/sync/.root.history
rm -fr ../data/users/test/*

echo "create folder1..."
ID1=`./ui_createfolder.sh folder1 | jq -r .ID`
laki.sh ls

echo "create folder2.."
ID2=`./ui_createfolder.sh folder2 $ID1 | jq -r .ID`
laki.sh find /

echo "create folder3.."
ID3=`./ui_createfolder.sh folder3 $ID1 | jq -r .ID`
laki.sh find /

echo "rename folder1 to other1...$ID1..."
./ui_rename.sh $ID1 other1
laki.sh find /

echo "delete folder2...$ID2"
./ui_delete.sh $ID2
laki.sh find /

echo "upload doc..."
DOCID=`./ui_upload.sh test.pdf | jq -r '.[].ID'`

echo "download doc..."
./ui_download.sh $DOCID

echo "done"
