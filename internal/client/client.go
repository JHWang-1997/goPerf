package client

import (
	"bufio"
	"crypto/tls"
	"goPerf/internal/model"
	"io"
	"net/http"
	"strings"
	"time"
)

type HandleResp func()

type Client interface {
	DoSend(request *model.Request, ch chan *model.Report)
}

type HttpClient struct {
	client0 *http.Client
}

func (h *HttpClient) DoSend(request *model.Request, ch chan *model.Report) {
	report := &model.Report{}
	start := time.Now()
	url := request.Url
	body := request.Body
	method := request.Method
	headers := request.HttpHeader
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		report.StatCode = 500
	}
	for key, val := range headers {
		req.Header.Add(key, val)
	}
	client := h.getClient(request.Keepalive, request.ReadTimeout)
	resp, err := client.Do(req)
	if err != nil {
		report.StatCode = 500
	}
	if "sse" == request.Protocol {
		err = h.handleSse(resp, report, start)
	} else {
		err = h.handleResp(resp, report, start)
	}
	report.Delay = uint64(time.Now().Sub(start).Nanoseconds())
	if err != nil && err != io.EOF {
		report.StatCode = 500
	}
	ch <- report
}

func (h *HttpClient) handleSse(resp *http.Response, report *model.Report, start time.Time) error {
	report.StatCode = resp.StatusCode
	body := resp.Body
	defer body.Close()
	first := true
	reader := bufio.NewReader(body)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			// 读完了
			return err
		}
		if strings.TrimRight(line, "\n\r") == "" {
			// 当前事件结束
			if first {
				first = false
				report.FirstTokenDelay = uint64(time.Now().Sub(start).Nanoseconds())
			} else {
				report.FollowTokenCount++
			}
		}
		report.ResBodyLen += uint64(len(line))
	}
}

func (h *HttpClient) handleResp(resp *http.Response, report *model.Report, start time.Time) error {
	report.StatCode = resp.StatusCode
	body := resp.Body
	defer body.Close()
	reader := bufio.NewReader(body)
	content, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	report.ResBodyLen = uint64(len(content))
	return nil
}

func (h *HttpClient) getClient(keepalive bool, timeout int) *http.Client {
	if h.client0 != nil {
		return h.client0
	}
	tr := &http.Transport{
		// 忽略服务端证书校验
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr, Timeout: time.Duration(timeout) * time.Second}
	if keepalive {
		h.client0 = client
	}
	return client
}
