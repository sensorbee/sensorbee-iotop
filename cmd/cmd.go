package cmd

import (
	"fmt"
	"pfi/sensorbee/dimo-sensorbee/server/config"

	"github.com/sensorbee/sensorbee-iotop/iotop"
	"gopkg.in/urfave/cli.v1"
)

func SetUp() cli.Command {
	cmd := cli.Command{
		Name:        "iotop",
		Usage:       "Monitoring tool for node I/O",
		Description: "iotop command launches a view for topology nodes",
		Action:      iotop.Run,
	}
	cmd.Flags = CmdFlags
	return cmd
}

// CmdFlags is list of command options.
var CmdFlags = []cli.Flag{
	cli.StringFlag{
		Name:   "uri",
		Value:  fmt.Sprintf("http://localhost:%d/", config.DefaultPort),
		Usage:  "the address of the target SensorBee server",
		EnvVar: "SENSORBEE_URI",
	},
	cli.StringFlag{
		Name:  "api-version",
		Value: "v1",
		Usage: "target API version",
	},
	cli.Float64Flag{
		Name:  "d",
		Value: 5.,
		Usage: "interval time [sec]",
	},
	cli.StringFlag{
		Name:  "topology,t",
		Usage: "the SensorBee topology to use",
	},
}
