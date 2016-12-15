package main

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

type generalLine struct {
	tplName  string
	name     string
	nodeType string
	state    string
}

type sourceLine struct {
	*generalLine
	out string
}

type boxLine struct {
	*generalLine
	processingTime string
	inOut          string
}

type sinkLine struct {
	*generalLine
	in string
}
