package iotop

import (
	"errors"
	"fmt"
	"strings"
	"time"

	cli "gopkg.in/urfave/cli.v1"
)

// MonitoringState is a global configuration on monitoring edge node I/O status.
type MonitoringState struct {
	d        time.Duration
	absFlag  bool
	hideEdge bool
	hideSrc  bool
	hideBox  bool
	hideSink bool
}

// SetUpMonitoringState sets up each configuration parameters.
func SetUpMonitoringState(c *cli.Context) (*MonitoringState, error) {
	d := c.Float64("d")
	if d < 1.0 {
		return nil, fmt.Errorf("interval must be over then 1[sec]")
	}

	absFlag := c.Bool("c")
	ms := &MonitoringState{
		d:       time.Duration(d*1000) * time.Millisecond,
		absFlag: absFlag,
	}
	if err := ms.setUpHideNodeLines(c.String("u")); err != nil {
		return nil, fmt.Errorf("invalid node name ('%v')", err)
	}

	return ms, nil
}

func (ms *MonitoringState) setUpHideNodeLines(visNode string) error {
	if visNode == "" {
		return nil
	}
	hideEdge := true
	hideSrc := true
	hideBox := true
	hideSink := true
	for _, n := range strings.Split(visNode, ",") {
		switch node := strings.TrimSpace(n); node {
		case "edge":
			hideEdge = false
		case "source", "src":
			hideSrc = false
		case "box":
			hideBox = false
		case "sink":
			hideSink = false
		default:
			return errors.New(node)
		}
	}
	ms.hideEdge = hideEdge
	ms.hideSrc = hideSrc
	ms.hideBox = hideBox
	ms.hideSink = hideSink
	return nil
}
