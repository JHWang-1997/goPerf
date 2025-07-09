package model

import (
	"fmt"
	"goPerf/pkg/utils"
	"strings"
)

var PROTOCOL = map[string]bool{
	"HTTP": true,
	"SSE":  true,
}

var OPERATION = map[string]bool{
	"POST":   true,
	"GET":    true,
	"PUT":    true,
	"DELETE": true,
}

type Request struct {
	Url          string            // 压测地址
	Method       string            // 请求类型
	HttpHeader   map[string]string // Http请求头
	Body         string            // 请求报文
	ReadTimeout  int               // 读取超时
	ConnTimeout  int               // 连接超时
	WriteTimeout int               // 发送超时
	Keepalive    bool              // 是否长连接
	Protocol     string            //协议类型
}

func (r *Request) SetProtocol(p string) {
	if !PROTOCOL[strings.ToUpper(p)] {
		panic("only support http or sse!")
	}
	r.Protocol = p
}

func (r *Request) SetKeepAlive(keepAlive bool) {
	r.Keepalive = keepAlive
}

func (r *Request) SetWriteTimeout(val string) {
	timeout := utils.ParseInt(val, "writeTimeout")
	if timeout < 0 {
		panic("write timeout can not below 0!")
	}
	r.WriteTimeout = timeout
}

func (r *Request) SetConnTimeout(val string) {
	timeout := utils.ParseInt(val, "connTimeout")
	if timeout < 0 {
		panic("connection timeout can not below 0!")
	}
	r.ConnTimeout = timeout
}

func (r *Request) SetReadTimeout(val string) {
	timeout := utils.ParseInt(val, "readTimeout")
	if timeout < 0 {
		panic("read timeout can not below 0!")
	}
	r.ReadTimeout = timeout
}

func (r *Request) SetBody(body string) {
	r.Body = body
}

func (r *Request) AddHeader(key string, val string) {
	if r.HttpHeader == nil {
		r.HttpHeader = make(map[string]string)
	}
	r.HttpHeader[key] = val
}

func (r *Request) SetUrl(url string) {
	if len(url) == 0 || !strings.HasPrefix(url, "http") {
		panic(fmt.Errorf("url is invalidate! url: %s", url))
	}
	r.Url = url
}

func (r *Request) SetOperation(op string) {
	if _, exist := OPERATION[strings.ToUpper(op)]; !exist {
		panic(fmt.Errorf("operation %s is not support!", op))
	}
	r.Method = op
}
