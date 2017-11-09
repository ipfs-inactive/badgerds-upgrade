#!/usr/bin/env bash

BADGER_08_HASH="QmTBxwy8cerzXbZQFUwTBCSxx55jUgVzudFcpmnAHUGuPy"
BADGER_10_HASH="Qme3aYm74r4gyMjtZTXpopHkBA6qjU21LXJy4GF56pLkW8"

PATCH_PATH=$(pwd)/patches
TEMP_PATH=$(mktemp -d)
cd ${TEMP_PATH}

do_work() {
  ipfs get $1
  cd $1/badger
  rm -rf .gx
  patch -p2 < ${PATCH_PATH}/$2
  gx publish

  echo '{'
  echo '    "author": "dgraph-io",'
  echo '    "hash": "'$(cat .gx/lastpubver | cut -d" " -f2)'",'
  echo '    "name": "badger",'
  echo '    "version": "'$(cat .gx/lastpubver | cut -d: -f1)'"'
  echo '}'
  cd ../..
}

do_work ${BADGER_08_HASH} badger08.patch
do_work ${BADGER_10_HASH} badger10.patch
