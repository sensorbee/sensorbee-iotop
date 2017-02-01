package iotop

import (
	"fmt"
	"strings"
	"time"
)

func hideNodeLines(ms *monitoringState, eb *editBox) (done struct{}) {
	done = struct{}{}

	in, err := eb.start("Which user (blank for all) ")
	eb.reset()
	if err != nil {
		eb.redrawAll(err.Error())
		<-time.After(2 * time.Second)
		return
	}
	if in == "" {
		ms.hideEdge = false
		ms.hideSrc = false
		ms.hideBox = false
		ms.hideSink = false
		return
	}
	hideEdge := true
	hideSrc := true
	hideBox := true
	hideSink := true
	for _, n := range strings.Split(in, ",") {
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
			eb.redrawAll(fmt.Sprintf("Invalid node name ('%v')", node))
			<-time.After(2 * time.Second)
			return
		}
	}
	ms.hideEdge = hideEdge
	ms.hideSrc = hideSrc
	ms.hideBox = hideBox
	ms.hideSink = hideSink
	return
}
