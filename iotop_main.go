package main

import (
	"os"

	"github.com/sensorbee/sensorbee-iotop/cmd"
	"github.com/sensorbee/sensorbee-iotop/iotop"

	"gopkg.in/urfave/cli.v1"
)

func main() {
	app := cli.NewApp()
	app.Name = "sensorbee iotop"
	app.Usage = "Monitoring tool for node I/O"
	app.Version = "0.0.1"
	app.Flags = cmd.CmdFlags
	app.Action = iotop.Run

	app.Run(os.Args)
}
