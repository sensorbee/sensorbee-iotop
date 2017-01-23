package iotop

import (
	"time"

	"gopkg.in/sensorbee/sensorbee.v0/data"
)

type nodeStatus struct {
	NodeName    string      `bql:"node_name"`
	NodeType    string      `bql:"node_type"`
	State       string      `bql:"state"`
	OutputStats outputStats `bql:"output_stats"`
	InputStats  inputStats  `bql:"input_stats"`
	Timestamp   time.Time   `bql:"ts"`
}

type outputStats struct {
	NumSentTotal int64    `bql:"num_sent_total`
	NumDropped   int64    `bql:"num_dropped"`
	Outputs      data.Map `bql:"outputs"`
}

type inputStats struct {
	NumReceivedTotal int64    `bql:"num_received_total`
	NumErrors        int64    `bql:"num_errors"`
	Inputs           data.Map `bql:"inputs"`
}

type sourcePipeStatus struct {
	NumQueued int64
	NumSent   int64
	QueueSize int64
}

type destinationPipeStatus struct {
	NumQueued   int64
	NumReceived int64
	QueueSize   int64
}

type generalLine struct {
	name     string
	nodeType string
	state    string
}

type sourceLine struct {
	*generalLine
	out     int64
	dropped int64
}

type boxLine struct {
	*generalLine
	processingTime int
	inOut          int64
	dropped        int64
	nerror         int64
}

type sinkLine struct {
	*generalLine
	in     int64
	nerror int64
}

type edgeLine struct {
	senderName        string
	senderNodeType    string
	receiverName      string
	receiverNodeType  string
	senderQueueSize   int64
	senderQueued      int64
	sent              int64
	receiverQueueSize int64
	receiverQueued    int64
	received          int64
	inOut             int64
}
