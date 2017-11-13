package upgrade

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	logging "log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	badger10 "gx/ipfs/QmQBccCGkYxLSdqzvUc6eTDqT9dqPcT7fCHzH6Z4ftWst3/badger"
	errors "gx/ipfs/QmVmDhyTTUcQXFD1rRQ64fGLMSAoaQvNH3hwuaCFAPq2hy/errors"
	lock "gx/ipfs/QmWi28zbQG6B1xfaaWx5cYoLn3kBFU6pQ6GWQNRV5P6dNe/lock"
	badger08 "gx/ipfs/QmaYHhxyszcAYob7WP8nSXnkJjzwfsWyApZEJFaJoJnXNP/badger"
)

var Log = logging.New(os.Stderr, "upgrade ", logging.LstdFlags)
var ErrInvalidVersion = errors.New("unsupported badger version")
var ErrCancelled = errors.New("context cancelled")

const (
	LockFile   = "repo.lock"
	ConfigFile = "config"
	SpecsFile  = "datastore_spec"

	SuppertedRepoVersion = 6
)

type keyValue struct {
	key   []byte
	value []byte
}

type Process struct {
	path string

	ctx    context.Context
	cancel context.CancelFunc

	dbPaths map[string]struct{}
}

func Upgrade(baseDir string) error {
	ctx, cancel := context.WithCancel(context.Background())
	p := Process{
		path: baseDir,

		ctx:    ctx,
		cancel: cancel,

		dbPaths: map[string]struct{}{},
	}

	err := p.checkRepoVersion()
	if err != nil {
		return err
	}

	unlock, err := lock.Lock(filepath.Join(p.path, LockFile))
	if err != nil {
		return err
	}
	defer unlock.Close()

	paths, err := p.loadSpecs()
	if err != nil {
		return err
	}

	for _, dir := range paths {
		err := p.upgradeDs(path.Join(p.path, dir))
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Process) upgradeDs(path string) error {
	Log.Printf("Upgrading badger at %s\n", path)

	Log.Printf("Trying badger 1.0\n")
	err := c.try10(path)
	if err == nil || err != ErrInvalidVersion {
		return err
	}

	Log.Printf("Trying badger 0.8\n")
	err = c.try08(path)
	if err == nil || err != ErrInvalidVersion {
		return err
	}

	return ErrInvalidVersion
}

func (c *Process) try10(path string) error {
	opt := badger10.DefaultOptions
	opt.Dir = path
	opt.ValueDir = path
	opt.SyncWrites = true

	db, err := badger10.Open(opt)
	if err != nil {
		if strings.HasPrefix(err.Error(), "manifest has unsupported version:") {
			err = ErrInvalidVersion
		}
		return err
	}

	db.Close()
	return nil
}

func (c *Process) try08(path string) error {
	opt := badger08.DefaultOptions
	opt.Dir = path
	opt.ValueDir = path
	opt.SyncWrites = true

	kv, err := badger08.NewKV(&opt)
	if err != nil {
		if strings.HasPrefix(err.Error(), "manifest has unsupported version:") {
			err = ErrInvalidVersion
		}
		return err
	}
	out := make(chan keyValue)
	go func() {
		opt := badger08.DefaultIteratorOptions
		opt.PrefetchValues = false
		it := kv.NewIterator(opt)
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			var kvi badger08.KVItem
			err := kv.Get(item.Key(), &kvi)
			if err != nil {
				Log.Printf("Error: %s\n", err.Error())
				it.Close()
				kv.Close()
				return
			}

			err = kvi.Value(func(d []byte) error {
				data := make([]byte, len(d))
				key := make([]byte, len(item.Key()))
				copy(data, d)
				copy(key, item.Key())

				select {
				case out <- keyValue{key: key, value: data}:
				case <-c.ctx.Done():
					return ErrCancelled
				}
				return nil
			})
			if err == ErrCancelled {
				it.Close()
				kv.Close()
				return
			}
			if err != nil {
				Log.Printf("Error: %s\n", err.Error())
				it.Close()
				kv.Close()
				return
			}
		}
		it.Close()
		kv.Close()

		close(out)
	}()

	return c.migrateData(out, path)
}

func (c *Process) migrateData(data chan keyValue, path string) error {
	temp, err := ioutil.TempDir(c.path, "badger-")
	if err != nil {
		c.cancel()
		return err
	}

	err = func() error {
		opt := badger10.DefaultOptions
		opt.ValueDir = temp
		opt.Dir = temp
		opt.SyncWrites = true
		db, err := badger10.Open(opt)
		if err != nil {
			c.cancel()
			return err
		}
		defer db.Close()

		Log.Printf("Moving data to %s\n", temp)
		n := 0

		var txn *badger10.Txn
		for entry := range data {
			if txn == nil {
				txn = db.NewTransaction(true)
			}

			err := txn.Set(entry.key, entry.value)
			if err != nil {
				c.cancel()
				txn.Discard()
				return err
			}

			if n%200 == 0 {
				err := txn.Commit(nil)
				if err != nil {
					return err
				}
				txn = nil
				Log.Printf("%d entries done\r\x1b[A", n)
			}
			n++
		}

		Log.Printf("%d entries done\n", n)
		Log.Printf("Commiting transaction\n")

		if txn != nil {
			return txn.Commit(nil)
		}
		return nil
	}()
	if err != nil {
		return err
	}

	backup, err := ioutil.TempDir(c.path, "badger-backup-")
	if err != nil {
		return err
	}
	if err = os.Remove(backup); err != nil {
		return err
	}

	Log.Printf("Renaming '%s' to '%s'\n", path, backup)

	if err = os.Rename(path, backup); err != nil {
		return err
	}
	Log.Printf("Renaming '%s' to '%s'\n", temp, path)

	if err = os.Rename(temp, path); err != nil {
		return err
	}

	Log.Printf("Success\n")
	Log.Printf("vvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvv")
	Log.Printf("AFTER YOU VERIFY THAT YOUR DATASTORE IS WORKING")
	Log.Printf("REMOVE '%s'", backup)
	Log.Printf("^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^")

	return nil
}

func (c *Process) loadSpecs() ([]string, error) {
	specData, err := ioutil.ReadFile(path.Join(c.path, SpecsFile))
	if err != nil {
		return nil, err
	}

	var spec map[string]interface{}
	err = json.Unmarshal(specData, &spec)
	if err != nil {
		return nil, err
	}

	return parseSpecs(spec)
}

func parseSpecs(spec map[string]interface{}) ([]string, error) {
	t, ok := spec["type"].(string)
	if !ok {
		return nil, errors.New("unexpected spec type")
	}

	switch t {
	case "mount":
		mounts, ok := spec["mounts"].([]interface{})
		if !ok {
			return nil, errors.New("unexpected mounts type")
		}

		var out []string

		for _, m := range mounts {
			mount, ok := m.(map[string]interface{})
			if !ok {
				return nil, errors.New("unexpected mount type")
			}

			paths, err := parseSpecs(mount)
			if err != nil {
				return nil, err
			}
			out = append(out, paths...)
		}
		return out, nil
	case "measure":
		child, ok := spec["child"].(map[string]interface{})
		if !ok {
			return nil, errors.New("unexpected child type")
		}

		return parseSpecs(child)
	case "badgerds":
		path, ok := spec["path"].(string)
		if !ok {
			return nil, errors.New("unexpected path type")
		}

		Log.Printf("Badger instance at %s\n", path)

		return []string{path}, nil
	case "flatfs", "levelds":
		return nil, nil
	default:
		return nil, errors.New("unexpected ds type")
	}
}

func (c *Process) checkRepoVersion() error {
	vstr, err := ioutil.ReadFile(filepath.Join(c.path, "version"))
	if err != nil {
		return err
	}

	version, err := strconv.Atoi(strings.TrimSpace(string(vstr)))
	if err != nil {
		return err
	}

	if version != SuppertedRepoVersion {
		return fmt.Errorf("unsupported fsrepo version: %d", version)
	}

	return nil
}
