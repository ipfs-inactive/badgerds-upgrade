#!/bin/bash

gx install
go build -o badgerds-upgrade

TEST_DIR=$(mktemp -d)

cp ./badgerds-upgrade ${TEST_DIR}
cd ${TEST_DIR}

wget https://dist.ipfs.io/go-ipfs/v0.4.12-rc2/go-ipfs_v0.4.12-rc2_linux-amd64.tar.gz
tar xzf go-ipfs_v0.4.12-rc2_linux-amd64.tar.gz
mv go-ipfs/ipfs ipfs0.4.12-rc2
rm -r go-ipfs

wget https://dist.ipfs.io/go-ipfs/v0.4.11/go-ipfs_v0.4.11_linux-amd64.tar.gz
tar xzf go-ipfs_v0.4.11_linux-amd64.tar.gz
mv go-ipfs/ipfs ipfs0.4.11
rm -r go-ipfs

mkdir ipfs
export IPFS_PATH=${TEST_DIR}/ipfs

./ipfs0.4.11 init --profile=badgerds,test

mkdir testfiles
for i in {1..10000}; do
  echo "$i" > "testfiles/$i"
done

./ipfs0.4.11 add testfiles --local -rqQ
./ipfs0.4.11 pin ls | wc -l

./badgerds-upgrade upgrade 2>&1

./ipfs0.4.12-rc2 pin ls | wc -l

cd ..
rm -rf ${TEST_DIR}
