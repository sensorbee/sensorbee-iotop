package iotop

import "time"

type monitoringState struct {
	d        time.Duration
	absFlag  bool
	hideEdge bool
	hideSrc  bool
	hideBox  bool
	hideSink bool
}
