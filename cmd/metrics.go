package app

import (
	"context"
	"goPerf/internal/model"
	"strings"
	"sync"
	"time"
)

var mutex = sync.Mutex{}

func HandleReport(ctx context.Context, wg *sync.WaitGroup, conf *model.PerfConf, request *model.Request, statCh chan *model.Report) {
	defer wg.Done()
	// 通知周期报告协程关闭
	stopCh := make(chan bool)
	// 获取校验器
	verify := model.GetVerify(conf)
	// 开启周期报告
	stat := getStat(request.Protocol)
	go periodPrint(&stat, time.Duration(conf.ReportTime), stopCh)

	// 处理请求的上报信息
	for {
		select {
		case <-ctx.Done():
			// fmt.Println("ctrl + C 结束进程")
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
	now := time.Now()
	stopCh <- true
	stat.ReportLast(now)
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
	} else {
		return &model.HttpStat{StartTime: time.Now()}
	}
}
