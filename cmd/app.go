package app

import (
	"context"
	"fmt"
	"goPerf/internal/client"
	"goPerf/internal/model"
	"sync"
)

func Start(ctx context.Context, request *model.Request, conf *model.PerfConf) {
	printTitle(request.Protocol)
	var waitGroup = &sync.WaitGroup{}
	statCh := make(chan *model.Report, 5000)
	// 启动指标统计
	waitGroup.Add(1)
	go HandleReport(ctx, waitGroup, conf, request, statCh)
	concurrent := conf.Concurrent
	for i := 1; i <= concurrent; i++ {
		c := &client.HttpClient{}
		go loop(ctx, waitGroup, statCh, request, c)
	}
	waitGroup.Wait()
}

func printTitle(protocol string) {
	if "http" == protocol {
		//fmt.Println("__________________________________________________________________________________________")
		fmt.Println("│****************************************************************************************│")
		fmt.Println("│******************* This is goPerf v1.0   Author: J-H.Wang w-------- *******************│")
		fmt.Println("│****************************************************************************************│")
		fmt.Println("│----------------------------------------------------------------------------------------│")
		fmt.Printf("│%10s  │  %-7s %-7s │   %-7s %-7s│   %-6s %-6s│   %-8s %-8s│\n", "runtime", "tps0", "delay0", "tps", "deley", "min", "max", "total", "success")
		fmt.Println("│------------┼------------------┼------------------┼----------------┼--------------------┼")

	} else if "sse" == protocol {
		//fmt.Println("__________________________________________________________________________________________")
		fmt.Println("│********************************************************************************************│")
		fmt.Println("│********************* This is goPerf v1.0   Author: J-H.Wang w-------- *********************│")
		fmt.Println("│********************************************************************************************│")
		fmt.Println("│--------------------------------------------------------------------------------------------│")
		fmt.Printf("│%9s  │   %-7s %-7s    %-7s %-6s │   %3s  %-5s  │   %-8s %-8s│\n",
			"runtime", "1'TD", "1'maxTD", "tokenTD", "token/s", "AllTD", "maxAllTD", "total", "success")

	}

}

func loop(ctx context.Context, group *sync.WaitGroup, statCh chan *model.Report, request *model.Request, client *client.HttpClient) {
	for {
		if ctx.Err() != nil {
			group.Done()
			return
		}
		_ = client.DoSend(request, statCh)
	}
}
