package app

import (
	"context"
	"fmt"
	"goPerf/internal/client"
	"goPerf/internal/model"
	"strings"
	"sync"
)

func Start(ctx context.Context, request *model.Request, conf *model.PerfConf) {
	printTitle(request.Protocol)
	var wg sync.WaitGroup
	statCh := make(chan *model.Report, 5000)

	// 统计协程：消费 statCh，周期/最终报告
	wg.Add(1)
	go func() {
		defer wg.Done()
		HandleReport(ctx, conf, request, statCh)
	}()

	// 压测协程：按并发数启动 worker，每个 worker 循环发请求并上报到 statCh
	concurrent := conf.Concurrent
	for i := 0; i < concurrent; i++ {
		wg.Add(1)
		c := &client.HttpClient{}
		go func() {
			defer wg.Done()
			workerLoop(ctx, statCh, request, c)
		}()
	}

	wg.Wait()
}

func printTitle(protocol string) {
	switch strings.ToLower(protocol) {
	case "http":
		fmt.Println("│****************************************************************************************│")
		fmt.Println("│******************* This is goPerf v1.0   Author: J-H.Wang *******************│")
		fmt.Println("│****************************************************************************************│")
		fmt.Println("│----------------------------------------------------------------------------------------│")
		fmt.Printf("│%10s  │  %-7s %-7s │   %-7s %-7s│   %-6s %-6s│   %-8s %-8s│\n", "runtime", "tps0", "delay0", "tps", "delay", "min", "max", "total", "success")
		fmt.Println("│------------┼------------------┼------------------┼----------------┼--------------------┼")
	case "sse":
		fmt.Println("│********************************************************************************************│")
		fmt.Println("│********************* This is goPerf v1.0   Author: J-H.Wang *********************│")
		fmt.Println("│********************************************************************************************│")
		fmt.Println("│--------------------------------------------------------------------------------------------│")
		fmt.Printf("│%9s  │   %-7s %-7s    %-7s %-6s │   %3s  %-5s  │   %-8s %-8s│\n",
			"runtime", "1'TD", "1'maxTD", "tokenTD", "token/s", "AllTD", "maxAllTD", "total", "success")
	default:
		fmt.Println("│ goPerf v1.0 │ protocol:", protocol)
	}
}

func workerLoop(ctx context.Context, statCh chan *model.Report, request *model.Request, c *client.HttpClient) {
	for ctx.Err() == nil {
		c.DoSend(request, statCh)
	}
}
