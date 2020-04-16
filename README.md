## This repository has been archived!

*This IPFS-related repository has been archived, and all issues are therefore frozen*. If you want to ask a question or open/continue a discussion related to this repo, please visit the [official IPFS forums](https://discuss.ipfs.io).

We archive repos for one or more of the following reasons:

- Code or content is unmaintained, and therefore might be broken
- Content is outdated, and therefore may mislead readers
- Code or content evolved into something else and/or has lived on in a different place
- The repository or project is not active in general

Please note that in order to keep the primary IPFS GitHub org tidy, most archived repos are moved into the [ipfs-inactive](https://github.com/ipfs-inactive) org.

If you feel this repo should **not** be archived (or portions of it should be moved to a non-archived repo), please [reach out](https://ipfs.io/help) and let us know. Archiving can always be reversed if needed.

---
   
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

### Upgrade badger datastores to latest supported version
```
$ badgerds-upgrade upgrade
```

Depending on your datastore size this might take a long time.
