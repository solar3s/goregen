package web

import (
	"fmt"
	"github.com/rkjdid/util"
	"github.com/solar3s/goregen/regenbox"
	"io/ioutil"
	"log"
	"path/filepath"
	"time"
)

type ChartLog struct {
	User          User
	Battery       Battery
	Resistor      util.Float
	CycleType     string
	TargetReached bool
	Reason        string
	TotalDuration util.Duration
	Config        regenbox.Config
	Measures      util.TimeSeries
	Measures1     util.TimeSeries
	Measures2     util.TimeSeries
	Measures3     util.TimeSeries
}

func (cl ChartLog) Info() ChartLogInfo {
	return ChartLogInfo{
		User:      cl.User,
		Battery:   cl.Battery,
		CycleType: cl.CycleType,
		StartTime: cl.Measures.Start,
		EndTime:   cl.Measures.End,
		Interval:  cl.Measures.Interval,
	}
}

func (cl ChartLog) FileName() string {
	return cl.Info().FileName()
}

func (cl ChartLog) String() string {
	return cl.Info().String()
}

type ChartLogInfo struct {
	User      User
	Battery   Battery
	CycleType string
	StartTime time.Time
	EndTime   time.Time
	Interval  util.Duration
	relPath   string
}

func (cli ChartLogInfo) String() string {
	return cli.Path()
}

func (cli ChartLogInfo) Path() string {
	if len(cli.relPath) > 0 {
		return cli.relPath
	}
	return cli.FileName()
}

func (cli ChartLogInfo) FileName() string {
	return fmt.Sprintf("%s_%s_%s_%s.log",
		cli.User.BetaId,
		cli.Battery.BetaRef,
		cli.CycleType,
		cli.StartTime.Format("2006-01-02_15h04m05"))
}

func ListChartLogs(dir string) (err error, infos []ChartLogInfo) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err, nil
	}
	for _, fi := range files {
		var fpath = filepath.Join(dir, fi.Name())
		var cl ChartLog
		err = util.ReadTomlFile(&cl, fpath)
		if err != nil {
			log.Printf("error parsing chart log: %s", err)
		} else {
			info := cl.Info()
			info.relPath = fi.Name()
			infos = append(infos, info)
		}
	}
	return nil, infos
}
