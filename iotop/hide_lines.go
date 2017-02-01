package iotop

import (
	"fmt"
	"time"
)

func hideNodeLines(ms *MonitoringState, eb *editBox) (done struct{}) {
	done = struct{}{}
	defer eb.reset()

	in, err := eb.start("Which user (blank for all) ")
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

	if err := ms.setUpHideNodeLines(in); err != nil {
		eb.redrawAll(fmt.Sprintf("Invalid node name ('%v')", err))
		<-time.After(2 * time.Second)
		return
	}
	return
}
