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
	NumQueued int
	NumSent   int
	QueueSize int
}

type destinationPipeStatus struct {
	NumQueued   int
	NumReceived int
	QueueSize   int
}

type generalLine struct {
	tplName  string
	name     string
	nodeType string
	state    string
}

type sourceLine struct {
	*generalLine
	out     string
	dropped string
}

type boxLine struct {
	*generalLine
	processingTime string
	inOut          string
	dropped        string
	nerror         string
}

type sinkLine struct {
	*generalLine
	in     string
	nerror string
}

type edgeLine struct {
	senderName        string
	senderNodeType    string
	receiverName      string
	receiverNodeType  string
	senderQueueSize   string
	senderQueued      string
	sent              string
	sentInt           int
	receiverQueueSize string
	receiverQueued    string
	received          string
	inOut             string
}
