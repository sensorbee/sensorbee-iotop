package iotop

import (
	"fmt"
	"strconv"
	"time"
)

func updateInterval(ms *MonitoringState, eb *editBox) (done struct{}) {
	done = struct{}{}
	defer eb.reset()

	in, err := eb.start(fmt.Sprintf("Change delay from %v to ", ms.d))
	if err != nil {
		eb.redrawAll(err.Error())
		<-time.After(2 * time.Second)
		return
	}
	if in == "" {
		return
	}
	interval, err := strconv.ParseFloat(in, 64)
	if err != nil {
		eb.redrawAll(fmt.Sprintf("Unacceptable floating point, %v", err))
		<-time.After(2 * time.Second)
		return
	}
	ms.d = time.Duration(interval*1000) * time.Millisecond
	return
}
