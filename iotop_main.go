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

const iotopTerminalColor = termbox.ColorDefault

func draw(c *cli.Context) {
	termbox.Clear(iotopTerminalColor, iotopTerminalColor)

	// test
	msgs, err := getNodeStatus(c)
	if err != nil {
		msgs = err.Error()
	}
	for i, msg := range strings.Split(msgs, "\n") {
		tbprint(0, i, iotopTerminalColor, iotopTerminalColor, msg)
	}
	termbox.Flush()
}

func tbprint(x, y int, fg, bg termbox.Attribute, msg string) {
	for _, c := range msg {
		termbox.SetCell(x, y, c, fg, bg)
		x += runewidth.RuneWidth(c)
	}
}

var (
	outputStatsNSentPath    = data.MustCompilePath("output_stats.num_sent_total")
	outputStatsNDroppedPath = data.MustCompilePath("output_stats.num_dropped")
	outputStatsOutputPath   = data.MustCompilePath("output_stats.outputs")
	inputStatsNRcvPath      = data.MustCompilePath("input_stats.num_received_total")
	inputStatsNErrPath      = data.MustCompilePath("input_stats.num_errors")
	inputStatsInputPath     = data.MustCompilePath("input_stats.inputs")
	dataMapDecloder         = data.NewDecoder(nil)
)

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
	edges := map[string]*edgeLine{}
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
			if out, err := sn.Status.Get(outputStatsNSentPath); err == nil {
				outNum, _ := data.ToInt(out)
				line.out = fmt.Sprintf("%d", outNum)
			}
			if dropped, err := sn.Status.Get(outputStatsNDroppedPath); err == nil {
				droppedNum, _ := data.ToInt(dropped)
				line.dropped = fmt.Sprintf("%d", droppedNum)
			}
			srcs = append(srcs, line)

			if outputs, err := sn.Status.Get(outputStatsOutputPath); err == nil {
				outputMap, _ := data.AsMap(outputs)
				for outName, output := range outputMap {
					om, _ := data.AsMap(output)
					line := getSourcePipeStatus(sn.Name, sn.NodeType, om)
					line.receiverName = outName
					key := fmt.Sprintf("%s|%s", sn.Name, outName)
					edges[key] = line
				}
			}
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
			if in, err := bn.Status.Get(inputStatsNRcvPath); err == nil {
				if out, err := bn.Status.Get(outputStatsNSentPath); err == nil {
					inNum, _ := data.ToInt(in)
					outNum, _ := data.ToInt(out)
					line.inOut = fmt.Sprintf("%d", outNum-inNum)
				}
			}
			// TODO: process time
			if dropped, err := bn.Status.Get(outputStatsNDroppedPath); err == nil {
				droppedNum, _ := data.ToInt(dropped)
				line.dropped = fmt.Sprintf("%d", droppedNum)
			}
			if nerror, err := bn.Status.Get(inputStatsNErrPath); err == nil {
				nerrorNum, _ := data.ToInt(nerror)
				line.nerror = fmt.Sprintf("%d", nerrorNum)
			}
			boxes = append(boxes, line)

			if inputs, err := bn.Status.Get(inputStatsInputPath); err == nil {
				inputMap, _ := data.AsMap(inputs)
				for inName, input := range inputMap {
					line, ok := edges[fmt.Sprintf("%s|%s", inName, bn.Name)]
					if !ok {
						continue // TODO: should not be ignored
					}
					im, _ := data.AsMap(input)
					addDestinationPipeStatus(bn.NodeType, im, line)
				}
			}
			if outputs, err := bn.Status.Get(outputStatsOutputPath); err == nil {
				outputMap, _ := data.AsMap(outputs)
				for outName, output := range outputMap {
					om, _ := data.AsMap(output)
					line := getSourcePipeStatus(bn.Name, bn.NodeType, om)
					line.receiverName = outName
					key := fmt.Sprintf("%s|%s", bn.Name, outName)
					edges[key] = line
				}
			}
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
			if in, err := sn.Status.Get(inputStatsNRcvPath); err == nil {
				inNum, _ := data.ToInt(in)
				line.in = fmt.Sprintf("%d", inNum)
			}
			if nerror, err := sn.Status.Get(inputStatsNErrPath); err == nil {
				nerrorNum, _ := data.ToInt(nerror)
				line.nerror = fmt.Sprintf("%d", nerrorNum)
			}
			sinks = append(sinks, line)

			if inputs, err := sn.Status.Get(inputStatsInputPath); err == nil {
				inputMap, _ := data.AsMap(inputs)
				for inName, input := range inputMap {
					line, ok := edges[fmt.Sprintf("%s|%s", inName, sn.Name)]
					if !ok {
						continue // TODO: should not be ignored
					}
					im, _ := data.AsMap(input)
					addDestinationPipeStatus(sn.NodeType, im, line)
				}
			}
		}
	}

	b := bytes.NewBuffer(nil)
	w := tabwriter.NewWriter(b, 0, 0, 1, ' ', 0)
	fmt.Fprintln(w, "SENDER\tSTYPE\tRCVER\tRTYPE\tSQSIZE\tSQNUM\tSNUM\tRQSIZE\tRQNUM\tRNUM\tINOUT")
	for _, l := range edges {
		values := fmt.Sprintf("%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t",
			l.senderName, l.senderNodeType, l.receiverName, l.receiverNodeType,
			l.senderQueueSize, l.senderQueued, l.sent,
			l.receiverQueueSize, l.receiverQueued, l.received, l.inOut)
		fmt.Fprintln(w, values)
	}
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "TPLGY\tNAME\tNTYPE\tSTATE\tOUT\tDROP")
	for _, l := range srcs {
		values := fmt.Sprintf("%v\t%v\t%v\t%v\t%v\t%v",
			l.tplName, l.name, l.nodeType, l.state, l.out, l.dropped)
		fmt.Fprintln(w, values)
	}
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "TPLGY\tNAME\tNTYPE\tSTATE\tINOUT\tDROP\tERR")
	for _, l := range boxes {
		values := fmt.Sprintf("%v\t%v\t%v\t%v\t%v\t%v\t%v",
			l.tplName, l.name, l.nodeType, l.state, l.inOut, l.dropped, l.nerror)
		fmt.Fprintln(w, values)
	}
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "TPLGY\tNAME\tNTYPE\tSTATE\tIN\tERR")
	for _, l := range sinks {
		values := fmt.Sprintf("%v\t%v\t%v\t%v\t%v\t%v",
			l.tplName, l.name, l.nodeType, l.state, l.in, l.nerror)
		fmt.Fprintln(w, values)
	}

	w.Flush()
	return b.String(), nil
}

func getSourcePipeStatus(name, nodeType string, output data.Map) *edgeLine {
	line := &edgeLine{
		senderName:     name,
		senderNodeType: nodeType,
	}
	pipeSts := &sourcePipeStatus{}
	if err := dataMapDecloder.Decode(output, pipeSts); err != nil {
		return line
	}
	line.senderQueued = fmt.Sprintf("%d", pipeSts.NumQueued)
	line.senderQueueSize = fmt.Sprintf("%d", pipeSts.QueueSize)
	line.sent = fmt.Sprintf("%d", pipeSts.NumSent)
	line.sentInt = pipeSts.NumSent
	return line
}

func addDestinationPipeStatus(nodeType string, input data.Map, line *edgeLine) {
	line.receiverNodeType = nodeType
	pipeSts := &destinationPipeStatus{}
	if err := dataMapDecloder.Decode(input, pipeSts); err != nil {
		return
	}
	line.receiverQueued = fmt.Sprintf("%d", pipeSts.NumQueued)
	line.receiverQueueSize = fmt.Sprintf("%d", pipeSts.QueueSize)
	line.received = fmt.Sprintf("%d", pipeSts.NumReceived)
	line.inOut = fmt.Sprintf("%d", pipeSts.NumReceived-line.sentInt)
}
