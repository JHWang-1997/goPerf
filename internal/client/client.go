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
	defer func() {
		report.Delay = uint64(time.Since(start).Nanoseconds())
		ch <- report
	}()

	req, err := http.NewRequest(request.Method, request.Url, strings.NewReader(request.Body))
	if err != nil {
		report.StatCode = 500
		return
	}
	for k, v := range request.HttpHeader {
		req.Header.Add(k, v)
	}
	client := h.getClient(request.Keepalive, request.ReadTimeout)
	resp, err := client.Do(req)
	if err != nil {
		report.StatCode = 500
		return
	}
	if resp != nil {
		defer resp.Body.Close()
	}
	if strings.EqualFold(request.Protocol, "sse") {
		err = h.handleSse(resp, report, start)
	} else {
		err = h.handleResp(resp, report, start)
	}
	if err != nil && err != io.EOF {
		report.StatCode = 500
	}
}

func (h *HttpClient) handleSse(resp *http.Response, report *model.Report, start time.Time) error {
	if resp != nil {
		report.StatCode = resp.StatusCode
	}
	if resp == nil || resp.Body == nil {
		return nil
	}
	body := resp.Body
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
	if resp != nil {
		report.StatCode = resp.StatusCode
	}
	if resp == nil || resp.Body == nil {
		return nil
	}
	body := resp.Body
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
