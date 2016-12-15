package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/sensorbee/sensorbee.v0/client"

	"github.com/mattn/go-runewidth"
	"github.com/nsf/termbox-go"
	"gopkg.in/sensorbee/sensorbee.v0/server/config"
	"gopkg.in/urfave/cli.v1"
)

func main() {
	app := cli.NewApp()
	app.Name = "sensorbee iotop"
	app.Usage = "Monitoring tool for node I/O"
	app.Version = "0.0.1"
	app.Flags = CmdFlags
	app.Action = Run

	app.Run(os.Args)
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

// Run 'sensorbee-iotop' command.
func Run(c *cli.Context) error {
	// TODO: check os.Stdout().Fd() is terminal or not
	// ref: github.com/mattn/go-isatty

	d := c.Float64("d")
	if d < 1.0 {
		return fmt.Errorf("interval must be over than 1[sec]")
	}

	if err := termbox.Init(); err != nil {
		return fmt.Errorf("fail to initialize termbox, %v", err)
	}
	defer termbox.Close()

	eventQueue := make(chan termbox.Event)
	go func() {
		for {
			eventQueue <- termbox.PollEvent()
		}
	}()

	tick := time.Tick(time.Duration(d*1000) * time.Millisecond)
	go func() {
		for {
			draw(c)
			<-tick
		}
	}()

	running := true
	for running {
		select {
		case ev := <-eventQueue:
			if ev.Type == termbox.EventKey &&
				(ev.Key == termbox.KeyCtrlC || ev.Ch == 'q') {
				running = false
			}
		default:
			// nothing to do
		}
	}
	return nil
}

func draw(c *cli.Context) {
	const coldef = termbox.ColorDefault
	termbox.Clear(coldef, coldef)

	// test
	msgs, err := getNodeStatus(c)
	if err != nil {
		msgs = err.Error()
	}
	for i, msg := range strings.Split(msgs, "\n") {
		tbprint(0, i, coldef, coldef, msg)
	}
	termbox.Flush()
}

func tbprint(x, y int, fg, bg termbox.Attribute, msg string) {
	for _, c := range msg {
		termbox.SetCell(x, y, c, fg, bg)
		x += runewidth.RuneWidth(c)
	}
}

func getNodeStatus(c *cli.Context) (string, error) {
	res, err := do(c, client.Get, "node_status", nil, "cannot access node status")
	if err != nil {
		return "", err
	}

	var js map[string]interface{}
	res.ReadJSON(&js)
	fmtRes, err := json.MarshalIndent(js, "", "  ")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s", fmtRes), nil
}

func newRequester(c *cli.Context) (*client.Requester, error) {
	r, err := client.NewRequester(c.String("uri"), c.String("api-version"))
	if err != nil {
		return nil, fmt.Errorf("cannot create HTTP requester, %v", err)
	}
	return r, nil
}

func do(c *cli.Context, method client.Method, path string, body interface{},
	baseErrMsg string) (*client.Response, error) {
	req, err := newRequester(c)
	if err != nil {
		return nil, fmt.Errorf("%v, %v", baseErrMsg, err)
	}
	res, err := req.Do(method, path, body)
	if err != nil {
		return nil, fmt.Errorf("%v, %v", baseErrMsg, err)
	}
	if res.IsError() {
		errRes, err := res.Error()
		if err != nil {
			return nil, fmt.Errorf("%v and failed to parse error information, %v",
				baseErrMsg, err)
		}
		return nil, fmt.Errorf("%v: %v, %v: %v", baseErrMsg, errRes.Code,
			errRes.RequestID, errRes.Message)
	}
	return res, nil
}
