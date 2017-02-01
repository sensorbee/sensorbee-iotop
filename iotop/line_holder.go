package iotop

import (
	"bytes"
	"fmt"
	"sync"
	"text/tabwriter"
	"time"

	"gopkg.in/sensorbee/sensorbee.v0/data"
)

type prevLineHolder struct {
	srcs  map[string]sourceLine
	boxes map[string]boxLine
	sinks map[string]sinkLine
	edges map[string]*edgeLine
}

type lineHolder struct {
	rwm     sync.RWMutex
	srcs    map[string]sourceLine
	boxes   map[string]boxLine
	sinks   map[string]sinkLine
	edges   map[string]*edgeLine
	current time.Time
	prev    *prevLineHolder // not use lineHolder not to share other parameter
	decoder *data.Decoder
}

func newLineHolder() *lineHolder {
	prev := &prevLineHolder{
		srcs:  map[string]sourceLine{},
		boxes: map[string]boxLine{},
		sinks: map[string]sinkLine{},
		edges: map[string]*edgeLine{},
	}
	return &lineHolder{
		srcs:    map[string]sourceLine{},
		boxes:   map[string]boxLine{},
		sinks:   map[string]sinkLine{},
		edges:   map[string]*edgeLine{},
		current: time.Now(),
		prev:    prev,
		decoder: data.NewDecoder(nil),
	}
}

func (h *lineHolder) clear() { // hand GC to clear old maps
	h.srcs = map[string]sourceLine{}
	h.boxes = map[string]boxLine{}
	h.sinks = map[string]sinkLine{}
	h.edges = map[string]*edgeLine{}
}

func (h *lineHolder) push(m data.Map) error {
	h.rwm.Lock()
	defer h.rwm.Unlock()
	ns := &nodeStatus{}
	if err := h.decoder.Decode(m, ns); err != nil {
		return err
	}

	if h.current != ns.Timestamp {
		h.prev.srcs = h.srcs
		h.prev.boxes = h.boxes
		h.prev.sinks = h.sinks
		h.prev.edges = h.edges
		h.clear()
		h.current = ns.Timestamp
	}

	gl := &generalLine{
		name:     ns.NodeName,
		nodeType: ns.NodeType,
		state:    ns.State,
	}
	switch ns.NodeType {
	case "source":
		line := sourceLine{
			generalLine: gl,
			out:         ns.OutputStats.NumSentTotal,
			dropped:     ns.OutputStats.NumDropped,
		}
		h.srcs[ns.NodeName] = line
		h.setSourcePipeStatus(ns.NodeName, ns.NodeType, ns.OutputStats.Outputs)

	case "box":
		line := boxLine{
			generalLine: gl,
			inOut: ns.OutputStats.NumSentTotal -
				ns.InputStats.NumReceivedTotal,
			dropped: ns.OutputStats.NumDropped,
			nerror:  ns.InputStats.NumErrors,
		}
		// TODO: process time
		// TODO: BQL statement, when SELETE query
		h.boxes[ns.NodeName] = line
		h.setSourcePipeStatus(ns.NodeName, ns.NodeType, ns.OutputStats.Outputs)
		h.setDestinationPipeStatus(ns.NodeName, ns.NodeType, ns.InputStats.Inputs)

	case "sink":
		line := sinkLine{
			generalLine: gl,
			in:          ns.InputStats.NumReceivedTotal,
			nerror:      ns.InputStats.NumErrors,
		}
		h.sinks[ns.NodeName] = line
		h.setDestinationPipeStatus(ns.NodeName, ns.NodeType, ns.InputStats.Inputs)
	}
	return nil
}

func (h *lineHolder) setSourcePipeStatus(name, nodeType string, outputs data.Map) {
	if len(outputs) == 0 {
		return
	}
	for outName, output := range outputs {
		om, _ := data.AsMap(output)
		pipeSts := &sourcePipeStatus{}
		if err := h.decoder.Decode(om, pipeSts); err != nil {
			return
		}

		key := fmt.Sprintf("%s|%s", name, outName)
		line, ok := h.edges[key]
		if !ok {
			line = &edgeLine{}
			h.edges[key] = line
		} else {
			line.inOut = line.received - pipeSts.NumSent
		}
		line.senderName = name
		line.senderNodeType = nodeType

		line.senderQueued = pipeSts.NumQueued
		line.senderQueueSize = pipeSts.QueueSize
		line.sent = pipeSts.NumSent
	}
}

func (h *lineHolder) setDestinationPipeStatus(name, nodeType string, inputs data.Map) {
	if len(inputs) == 0 {
		return
	}
	for inName, input := range inputs {
		im, _ := data.AsMap(input)
		pipeSts := &destinationPipeStatus{}
		if err := h.decoder.Decode(im, pipeSts); err != nil {
			return
		}

		key := fmt.Sprintf("%s|%s", inName, name)
		line, ok := h.edges[key]
		if !ok {
			line = &edgeLine{}
			h.edges[key] = line
		} else {
			line.inOut = pipeSts.NumReceived - line.sent
		}
		line.receiverName = name
		line.receiverNodeType = nodeType

		line.receiverQueued = pipeSts.NumQueued
		line.receiverQueueSize = pipeSts.QueueSize
		line.received = pipeSts.NumReceived
	}
}

func (h *lineHolder) flush(ms *MonitoringState) string {
	h.rwm.RLock()
	defer h.rwm.RUnlock()
	b := bytes.NewBuffer(nil)
	w := tabwriter.NewWriter(b, 0, 0, 1, ' ', 0)

	if !ms.hideEdge {
		h.printEdgeLines(w, ms)
		fmt.Fprintln(w, "")
	}
	if !ms.hideSrc {
		h.printSrcLines(w, ms)
		fmt.Fprintln(w, "")
	}
	if !ms.hideBox {
		h.printBoxLines(w, ms)
		fmt.Fprintln(w, "")
	}
	if !ms.hideSink {
		h.printSinkLines(w, ms)
	}

	w.Flush()
	return b.String()
}

func (h *lineHolder) printEdgeLines(w *tabwriter.Writer, ms *MonitoringState) {
	fmt.Fprintln(w, "SENDER\tSTYPE\tRCVER\tRTYPE\tSQSIZE\tSQNUM\tSNUM\tRQSIZE\tRQNUM\tRNUM\tINOUT")
	for _, name := range edgeLineMap(h.edges).sortedKeys() {
		l := h.edges[name]
		var values string
		if prev, ok := h.prev.edges[name]; ok && !ms.absFlag {
			inout := float64(l.inOut-prev.inOut) / ms.d.Seconds()
			values = fmt.Sprintf("%v\t%v\t%v\t%v\t%d\t%d\t%d\t%d\t%d\t%d\t%.2f",
				l.senderName, l.senderNodeType, l.receiverName,
				l.receiverNodeType, l.senderQueueSize, l.senderQueued, l.sent,
				l.receiverQueueSize, l.receiverQueued, l.received, inout)
		} else {
			values = fmt.Sprintf("%v\t%v\t%v\t%v\t%d\t%d\t%d\t%d\t%d\t%d\t[%d]",
				l.senderName, l.senderNodeType, l.receiverName,
				l.receiverNodeType, l.senderQueueSize, l.senderQueued, l.sent,
				l.receiverQueueSize, l.receiverQueued, l.received, l.inOut)
		}
		fmt.Fprintln(w, values)
	}
}

func (h *lineHolder) printSrcLines(w *tabwriter.Writer, ms *MonitoringState) {
	fmt.Fprintln(w, "NAME\tNTYPE\tSTATE\tOUT\tDROP")
	for _, name := range sourceLineMap(h.srcs).sortedKeys() {
		l := h.srcs[name]
		var values string
		if prev, ok := h.prev.srcs[name]; ok && !ms.absFlag {
			out := float64(l.out-prev.out) / ms.d.Seconds()
			values = fmt.Sprintf("%v\t%v\t%v\t%.2f\t%d",
				l.name, l.nodeType, l.state, out, l.dropped)
		} else {
			values = fmt.Sprintf("%v\t%v\t%v\t[%d]\t%d",
				l.name, l.nodeType, l.state, l.out, l.dropped)
		}
		fmt.Fprintln(w, values)
	}
}

func (h *lineHolder) printBoxLines(w *tabwriter.Writer, ms *MonitoringState) {
	fmt.Fprintln(w, "NAME\tNTYPE\tSTATE\tINOUT\tDROP\tERR")
	for _, name := range boxLineMap(h.boxes).sortedKeys() {
		l := h.boxes[name]
		var values string
		if prev, ok := h.prev.boxes[name]; ok && !ms.absFlag {
			inout := float64(l.inOut-prev.inOut) / ms.d.Seconds()
			values = fmt.Sprintf("%v\t%v\t%v\t%.2f\t%d\t%d",
				l.name, l.nodeType, l.state, inout, l.dropped, l.nerror)
		} else {
			values = fmt.Sprintf("%v\t%v\t%v\t[%d]\t%d\t%d",
				l.name, l.nodeType, l.state, l.inOut, l.dropped, l.nerror)
		}
		fmt.Fprintln(w, values)
	}
}

func (h *lineHolder) printSinkLines(w *tabwriter.Writer, ms *MonitoringState) {
	fmt.Fprintln(w, "NAME\tNTYPE\tSTATE\tIN\tERR")
	for _, name := range sinkLineMap(h.sinks).sortedKeys() {
		l := h.sinks[name]
		var values string
		if prev, ok := h.prev.sinks[name]; ok && !ms.absFlag {
			in := float64(l.in-prev.in) / ms.d.Seconds()
			values = fmt.Sprintf("%v\t%v\t%v\t%.2f\t%d",
				l.name, l.nodeType, l.state, in, l.nerror)
		} else {
			values = fmt.Sprintf("%v\t%v\t%v\t[%d]\t%d",
				l.name, l.nodeType, l.state, l.in, l.nerror)
		}
		fmt.Fprintln(w, values)
	}
}
