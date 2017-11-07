package main

import (
	"os"
	"path"

	"github.com/ipfs/badgerds-upgrade/upgrade"

	"github.com/mitchellh/go-homedir"
	"gx/ipfs/QmVcLF2CgjQb5BWmYFWsDfxDjbzBfcChfdHRedxeL3dV4K/cli"
)

const (
	DefaultPathName   = ".ipfs"
	DefaultPathRoot   = "~/" + DefaultPathName
	DefaultConfigFile = "config"
	EnvDir            = "IPFS_PATH"
	Version           = "0.0.1"
)

func main() {
	app := cli.NewApp()

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "verbose",
			Usage: "print verbose logging information",
		},
	}

	app.Commands = []cli.Command{
		UpgradeCommand,
	}

	if err := app.Run(os.Args); err != nil {
		upgrade.Log.Fatal(err)
	}
}

var UpgradeCommand = cli.Command{
	Name:  "upgrade",
	Usage: "upgrade badger instances",
	Flags: []cli.Flag{

	},
	Action: func(c *cli.Context) error {
		baseDir, err := getBaseDir()
		if err != nil {
			upgrade.Log.Fatal(err)
		}

		err = upgrade.Upgrade(baseDir)
		if err != nil {
			upgrade.Log.Fatal(err)
		}
		return err
	},
}

func getBaseDir() (string, error) {
	baseDir := os.Getenv(EnvDir)
	if baseDir == "" {
		baseDir = DefaultPathRoot
	}

	baseDir, err := homedir.Expand(baseDir)
	if err != nil {
		return "", err
	}

	configFile := path.Join(baseDir, DefaultConfigFile)

	_, err = os.Stat(configFile)
	if err != nil {
		return "", err
	}

	return baseDir, nil
}
