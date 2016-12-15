package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"gopkg.in/sensorbee/sensorbee.v0/data"

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

	var status topologiesStatus
	res.ReadJSON(&status)

	srcs := []sourceLine{}
	boxes := []boxLine{}
	sinks := []sinkLine{}
	for _, tpl := range status.Topologies {
		// TODO: in/out unit should be [tuples/sec]

		for _, sn := range tpl.Souces {
			line := sourceLine{
				generalLine: &generalLine{
					tplName:  tpl.Name,
					name:     sn.Name,
					nodeType: sn.NodeType,
					state:    sn.State,
				},
			}
			if out, err := sn.Status.Get(data.MustCompilePath(
				"output_stats.num_sent_total")); err == nil {
				outNum, _ := data.ToInt(out)
				line.out = fmt.Sprintf("%d", outNum)
			}
			srcs = append(srcs, line)
		}

		for _, bn := range tpl.Boxes {
			line := boxLine{
				generalLine: &generalLine{
					tplName:  tpl.Name,
					name:     bn.Name,
					nodeType: bn.NodeType,
					state:    bn.State,
				},
			}
			if in, err := bn.Status.Get(data.MustCompilePath(
				"input_stats.num_received_total")); err == nil {
				if out, err := bn.Status.Get(data.MustCompilePath(
					"output_stats.num_sent_total")); err == nil {
					inNum, _ := data.ToInt(in)
					outNum, _ := data.ToInt(out)
					line.inOut = fmt.Sprintf("%d", outNum-inNum)
				}
			}
			// TODO: process time
			boxes = append(boxes, line)
		}

		for _, sn := range tpl.Sinks {
			line := sinkLine{
				generalLine: &generalLine{
					tplName:  tpl.Name,
					name:     sn.Name,
					nodeType: sn.NodeType,
					state:    sn.State,
				},
			}
			if in, err := sn.Status.Get(data.MustCompilePath(
				"input_stats.num_received_total")); err == nil {
				inNum, _ := data.ToInt(in)
				line.in = fmt.Sprintf("%d", inNum)
			}
			sinks = append(sinks, line)
		}
	}

	b := bytes.NewBuffer(nil)
	w := tabwriter.NewWriter(b, 0, 0, 1, ' ', 0)
	fmt.Fprintln(w, "TPLGY\tNAME\tNTYPE\tSTATE\tOUT")
	for _, l := range srcs {
		values := fmt.Sprintf("%v\t%v\t%v\t%v\t%v",
			l.tplName, l.name, l.nodeType, l.state, l.out)
		fmt.Fprintln(w, values)
	}
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "TPLGY\tNAME\tNTYPE\tSTATE\tINOUT")
	for _, l := range boxes {
		values := fmt.Sprintf("%v\t%v\t%v\t%v\t%v",
			l.tplName, l.name, l.nodeType, l.state, l.inOut)
		fmt.Fprintln(w, values)
	}
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "TPLGY\tNAME\tNTYPE\tSTATE\tIN")
	for _, l := range sinks {
		values := fmt.Sprintf("%v\t%v\t%v\t%v\t%v",
			l.tplName, l.name, l.nodeType, l.state, l.in)
		fmt.Fprintln(w, values)
	}

	w.Flush()
	return b.String(), nil
}
