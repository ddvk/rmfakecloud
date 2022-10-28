#!/bin/bash
# rudimentary smoke test
set -e

. ./common.env

rm -fr ../data/users/$USER/*
mkdir -p ../data/users/$USER/sync

export RMAPI_HOST=$URL
export RMAPI_CONFIG=.rmapi.config
CODE=`./ui_getcode.sh`
echo "code: $CODE"
#todo rmapi -ni
echo "create folder1..."
RESULT=`./ui_createfolder.sh folder1` 
ID1=`echo $RESULT | jq -r .ID`
rmapi ls

echo "create folder2.."
RESULT=`./ui_createfolder.sh folder2 $ID1`
ID2=`echo $RESULT | jq -r .ID`
rmapi find /

echo "create folder3.."
RESULT=`./ui_createfolder.sh folder3 $ID1`
ID3=`echo $RESULT | jq -r .ID`
rmapi find /

echo "rename folder1 to other1...$ID1..."
./ui_rename.sh $ID1 other1
rmapi ls

echo "delete folder2...$ID2"
./ui_delete.sh $ID2
rmapi find /

echo "upload doc..."
RESULT=`./ui_upload.sh test.pdf`
DOCID=`echo $RESULT | jq -r '.[].ID'`

echo "download doc..."
./ui_download.sh $DOCID
rmapi ls

echo "get documents"
./ui_documents.sh | jq



echo "done"
