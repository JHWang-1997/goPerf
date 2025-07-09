package model

import (
	"fmt"
	"goPerf/pkg/utils"
	"time"
)

type Verify func(report *Report) bool

type Stat interface {
	Add(report *Report, v Verify)
	ReportPeriod(duration time.Duration)
	ReportLast(finishTime time.Time)
	ResetPeriod()
}

type SseStat struct {
	StartTime        time.Time
	TotalNum         uint64
	SuccessNum       uint64
	TotalDelay       uint64
	MaxDelay         uint64
	MinDelay         uint64
	FirstTokenDelay  uint64
	MaxFirstDelay    uint64
	MinFirstDelay    uint64
	FollowTokenDelay uint64
	FollowTokenCount uint64
}

func (stat *SseStat) ResetPeriod() {

}

func (stat *SseStat) ReportLast(finishTime time.Time) {
	fmt.Println("\n|************************************** SUMMARY REPORT **************************************|")
	stat.ReportPeriod(0)
}

func (stat *SseStat) ReportPeriod(duration time.Duration) {
	var averFirstDelay, maxFirstDelay, averFollowDelay, tokenPs, averDelay, maxDelay uint64
	now := time.Now()
	sub := now.Sub(stat.StartTime)
	// 距离启动时间，精确到秒
	t := utils.Format(sub)
	// 最大首token时延（ms）
	maxFirstDelay = stat.MaxFirstDelay / 1e6
	if stat.FollowTokenCount > 0 {
		// 平均后续token时延(ms)
		averFollowDelay = stat.FollowTokenDelay / stat.FollowTokenCount / 1e6
	}
	// 平均每秒的token数
	tokenPs = uint64(stat.FollowTokenCount+stat.SuccessNum) * 1e9 / uint64(sub.Nanoseconds())
	if stat.SuccessNum > 0 {
		// 平均首token时延（ms）
		averFirstDelay = stat.FirstTokenDelay / stat.SuccessNum / 1e6
		// 平均总时延 (ms)
		averDelay = stat.TotalDelay / stat.SuccessNum / 1e6
	}
	// 最大总时延 (ms)
	maxDelay = stat.MaxDelay / 1e6
	info := fmt.Sprintf("|  %8s |%6d   %6d %10d%8d    | %6d%8d     |%6d  %8d    |",
		t, averFirstDelay, maxFirstDelay, averFollowDelay, tokenPs,
		averDelay, maxDelay, stat.TotalNum, stat.SuccessNum)
	fmt.Println(info)
}

func (stat *SseStat) Add(report *Report, v Verify) {
	stat.TotalNum++
	if !v(report) {
		return
	}
	stat.SuccessNum++
	stat.TotalDelay += report.Delay
	if report.Delay > stat.MaxDelay {
		stat.MaxDelay = report.Delay
	}
	if report.Delay < stat.MinDelay {
		stat.MinDelay = report.Delay
	}
	stat.FirstTokenDelay += report.FirstTokenDelay
	if report.FirstTokenDelay > stat.MaxFirstDelay {
		stat.MaxFirstDelay = report.FirstTokenDelay
	}
	if report.FirstTokenDelay < stat.MinFirstDelay {
		stat.MinFirstDelay = stat.FirstTokenDelay
	}
	report.FollowTokenDelay = report.Delay - report.FirstTokenDelay
	stat.FollowTokenDelay += report.FollowTokenDelay
	stat.FollowTokenCount += report.FollowTokenCount
}

type HttpStat struct {
	StartTime          time.Time
	TotalNum           uint64
	TotalDelay         uint64
	SuccessNum         uint64
	MinDelay           uint64
	MaxDelay           uint64
	PeriodSuccessCount uint64
	PeriodDelay        uint64
}

func (stat *HttpStat) ReportLast(finishTime time.Time) {
	sub := finishTime.Sub(stat.StartTime)
	//
	fmt.Println("******************************* SUMMARY REPORT*******************************")
	// runtime
	format := utils.Format(sub)
	// 计算平均tps
	averTps := stat.TotalNum * 1e9 / uint64(sub.Nanoseconds())
	// 计算平均时延 (ms)
	var averDelay uint64 = 0
	if stat.TotalNum > 0 {
		averDelay = stat.TotalDelay / stat.TotalNum / 1e6
	}
	fmt.Printf("runtime: %10s, total requet: %10d, successRequest: %10d, falureReuqest: %10d", format, stat.TotalNum, stat.SuccessNum, stat.TotalNum-stat.SuccessNum)
	fmt.Printf("averTps: %10d, averDelay: %10d", averTps, averDelay)
}

func (stat *HttpStat) ReportPeriod(duration time.Duration) {
	now := time.Now()
	sub := now.Sub(stat.StartTime)
	// 距离启动时间，精确到秒
	t := utils.Format(sub)
	// 周期内tps
	periodTps := stat.PeriodSuccessCount / uint64(duration)
	// 平均时延，毫秒
	var periodAverDelay uint64 = 0
	if stat.PeriodSuccessCount > 0 {
		periodAverDelay = stat.PeriodDelay / stat.PeriodSuccessCount / 1e6
	}
	// 统计tps
	tps := stat.TotalNum / uint64(sub.Seconds())
	// 统计平均时延(ms)
	var averDelay uint64 = 0
	if stat.TotalNum > 0 {
		averDelay = stat.TotalDelay / stat.TotalNum / 1e6
	}
	// 最大时延（ms）
	maxDelay := stat.MaxDelay / 1e6
	// 最小时延 (ms)
	minDelay := stat.MinDelay / 1e6
	report := fmt.Sprintf("│ %10s │   %-7d %-7d│   %-7d %-7d│   %-6d %-6d│   %-8d %-8d│",
		t, periodTps, periodAverDelay, tps, averDelay, minDelay, maxDelay, stat.TotalNum, stat.SuccessNum)
	fmt.Println(report)
}

func (stat *HttpStat) Add(report *Report, v Verify) {
	stat.TotalNum++
	if !v(report) {
		return
	}
	stat.TotalDelay += report.Delay
	stat.SuccessNum++
	if report.Delay < stat.MinDelay || stat.MinDelay == 0 {
		stat.MinDelay = report.Delay
	}
	if report.Delay > stat.MaxDelay || stat.MaxDelay == 0 {
		stat.MaxDelay = report.Delay
	}
	stat.PeriodSuccessCount++
	stat.PeriodDelay += report.Delay
}

func (stat *HttpStat) ResetPeriod() {
	stat.PeriodDelay = 0
	stat.PeriodSuccessCount = 0
}

type Report struct {
	StatCode         int
	Delay            uint64
	ReqBodyLen       uint64
	ResBodyLen       uint64
	FirstTokenDelay  uint64
	FollowTokenDelay uint64
	FollowTokenCount uint64
}
