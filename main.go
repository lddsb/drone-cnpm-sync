package main

import (
	"fmt"
	"github.com/urfave/cli"
	"os"
	"sync"
)

var wg = &sync.WaitGroup{}
var version = "20190520"


func main() {

	app := cli.NewApp()
	app.Name = "CNPM package sync plugin"
	app.Usage = "Sync not exists package version to cnpm(taobao mirror)"
	app.Copyright = "Copyright © 2019 lddsb️"
	app.Action = run
	app.Version = version
	app.Flags = []cli.Flag{
		cli.StringSliceFlag{
			Name: "paths",
			Usage: "the path to package.json",
			EnvVar: "PLUGIN_PATH",
		},
	}

	_ = app.Run(os.Args)
}

func run(c *cli.Context) {
	plugin := Plugin{
		Path: c.StringSlice("paths"),
	}

	if err := plugin.Exec(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
