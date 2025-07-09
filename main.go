package main

import (
	"context"
	"flag"
	"fmt"
	"goPerf/cmd"
	"goPerf/internal/model"
	"goPerf/pkg/utils"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"
)

func main() {
	request, conf := parse()
	ctx := context.Background()
	// 当设置运行时间时，时间到达后终止程序
	appTime := conf.AppTime
	if appTime > 0 {
		_, cancelWithTime := context.WithTimeout(ctx, time.Duration(appTime)*time.Second)
		defer cancelWithTime()
	}
	// 当执行ctrl + c时，终止程序并清理上下文
	ctx, cancelFunc := context.WithCancel(ctx)
	go func() {
		ch := make(chan os.Signal)
		signal.Notify(ch, syscall.SIGINT)
		// 当接收到SIGINT（ctrl + c）时，释放资源。通知所有的协程结束
		<-ch
		cancelFunc()
	}()
	app.Start(ctx, request, conf)
}

func parse() (*model.Request, *model.PerfConf) {
	fmt.Println(os.Args)
	path := os.Args[1]
	exist := utils.FileExist(path)
	if !exist {
		panic(fmt.Sprintf("config file %s is not exist!", path))
	}
	request := parseRequest(path)
	perfConf := parseFlag()
	return request, perfConf

}

func parseRequest(path string) *model.Request {
	if strings.HasSuffix(path, "properties") {
		properties, err := utils.ParseProperties(path)
		if err != nil {
			panic(err)
		}
		return createRequest(properties)
	}
	panic("can not found properties file!")
}

func createRequest(config *map[string]string) *model.Request {
	request := &model.Request{Keepalive: true}
	request.SetBody((*config)["body"])
	request.SetConnTimeout(getOrDefault(config, "connTimeout", "10"))
	request.SetReadTimeout(getOrDefault(config, "readTimeout", "60"))
	request.SetWriteTimeout(getOrDefault(config, "writeTimeout", "60"))
	request.SetOperation(getOrDefault(config, "method", "GET"))
	request.SetUrl((*config)["url"])
	request.SetKeepAlive("true" == getOrDefault(config, "keepalive", "true"))
	request.SetProtocol((*config)["protocol"])
	return request
}

func getOrDefault(config *map[string]string, key string, dft string) string {
	val, exist := (*config)[key]
	if exist {
		return val
	}
	return dft
}

func parseFlag() *model.PerfConf {
	arg := os.Args[2]
	perfCfg := model.PerfConf{Concurrent: 1}
	perfCfg.SetConcurrent(arg)
	flag.BoolVar(&perfCfg.Debug, "d", false, "debug mode")
	flag.Uint64Var(&perfCfg.AppTime, "t", 0, "app runtime")
	flag.IntVar(&perfCfg.CpuNum, "c", runtime.NumCPU(), "cpu number")
	flag.IntVar(&perfCfg.ReportTime, "rt", 1, "report period")
	if perfCfg.CpuNum > 0 {
		runtime.GOMAXPROCS(perfCfg.CpuNum)
	}
	return &perfCfg
}
