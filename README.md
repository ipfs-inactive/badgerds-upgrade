badgerds-upgrade
================

[![](https://img.shields.io/badge/made%20by-Protocol%20Labs-blue.svg?style=flat-square)](http://ipn.io)
[![](https://img.shields.io/badge/project-IPFS-blue.svg?style=flat-square)](http://ipfs.io/)
[![](https://img.shields.io/badge/freenode-%23ipfs-blue.svg?style=flat-square)](http://webchat.freenode.net/?channels=%23ipfs)

> Badger datastore upgrade tool

## Install

### Build From Source

These instructions assume that go has been installed as described [here](https://github.com/ipfs/go-ipfs#install-go).

```
$ go get -u github.com/ipfs/badgerds-upgrade
$ go get github.com/whyrusleeping/gx
$ go get github.com/whyrusleeping/gx-go
$ gx install
$ go install
```

## Usage

### Upgrade to latest supported version
```
$ badgerds-upgrade upgrade
```

Depending on your datastore size this might take a long time.