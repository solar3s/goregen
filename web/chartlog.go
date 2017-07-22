package web

import (
	"github.com/rkjdid/util"
	"github.com/solar3s/goregen/regenbox"
)

type ChartLog struct {
	User          User
	Battery       Battery
	CycleType     string
	TargetReached bool
	Reason        string
	TotalDuration util.Duration
	Config        regenbox.Config
	Measures      util.TimeSeries
}
