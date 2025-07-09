package model

import (
	"fmt"
	"goPerf/pkg/utils"
)

type PerfConf struct {
	Concurrent    int    // 并发数
	CpuNum        int    // 使用的CPU核数
	AppTime       uint64 // 系统运行时间 (s)
	Debug         bool   // debug模式？
	ReportTime    int    // 报告周期 (s)
	VerifyByCode  int    // 通过响应码校验
	VerifyByEvent int    // 通过事件个数校验

}

func (p *PerfConf) SetReportTime(time int) {
	if time < 0 {
		time = 1
	}
	p.ReportTime = time
}

func (p *PerfConf) SetCpuNum(num int) {
	p.CpuNum = num
}

func (p *PerfConf) SetConcurrent(arg string) {
	num := utils.ParseInt(arg, "concurrent")
	if num < 1 {
		panic(fmt.Errorf("concurrent number invalid! %s", arg))
	}
	p.Concurrent = num
}
