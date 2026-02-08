package app

import (
	"context"
	"goPerf/internal/model"
	"strings"
	"sync"
	"time"
)

var mutex = sync.Mutex{}

func HandleReport(ctx context.Context, conf *model.PerfConf, request *model.Request, statCh chan *model.Report) {
	stopCh := make(chan bool)
	verify := model.GetVerify(conf)
	stat := getStat(request.Protocol)
	go periodPrint(&stat, time.Duration(conf.ReportTime), stopCh)

	for {
		select {
		case <-ctx.Done():
			goto finish
		case info, ok := <-statCh:
			if !ok {
				goto finish
			}
			mutex.Lock()
			stat.Add(info, verify)
			mutex.Unlock()
		}
	}
finish:
	stopCh <- true
	stat.ReportLast(time.Now())
}

func periodPrint(stat *model.Stat, period time.Duration, stopCh chan bool) {
	ticker := time.NewTicker(period * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			mutex.Lock()
			(*stat).ReportPeriod(period)
			(*stat).ResetPeriod()
			mutex.Unlock()
		case <-stopCh:
			return
		}
	}
}

func getStat(p string) model.Stat {
	if strings.EqualFold("SSE", p) {
		return &model.SseStat{StartTime: time.Now()}
	}
	return &model.HttpStat{StartTime: time.Now()}
}
