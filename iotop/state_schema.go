package iotop

import (
	"gopkg.in/sensorbee/sensorbee.v0/server/response"
)

type topologiesStatus struct {
	Topologies []topologyStatus `json:"topologies"`
}

type topologyStatus struct {
	Name   string            `json:"name"`
	Souces []response.Source `json:"sources"`
	Boxes  []response.Stream `json:"boxes"`
	Sinks  []response.Sink   `json:"sinks"`
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
	tplName  string
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
	tplName           string
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
