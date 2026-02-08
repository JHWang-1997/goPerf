package config

import (
	"fmt"
	"goPerf/internal/model"
	"goPerf/pkg/utils"
	"strings"
)

// LoadRequest 从 .properties 文件加载请求配置，得到 *model.Request
func LoadRequest(path string) (*model.Request, error) {
	if path == "" {
		return nil, fmt.Errorf("config file path is empty")
	}
	if !utils.FileExist(path) {
		return nil, fmt.Errorf("config file %s does not exist", path)
	}
	if !strings.HasSuffix(path, ".properties") {
		return nil, fmt.Errorf("config file must be .properties: %s", path)
	}
	properties, err := utils.ParseProperties(path)
	if err != nil {
		return nil, err
	}
	return requestFromMap(properties), nil
}

func requestFromMap(m *map[string]string) *model.Request {
	cfg := *m
	req := &model.Request{Keepalive: true}
	req.SetBody(cfg["body"])
	req.SetConnTimeout(getOr(cfg, "connTimeout", "10"))
	req.SetReadTimeout(getOr(cfg, "readTimeout", "60"))
	req.SetWriteTimeout(getOr(cfg, "writeTimeout", "60"))
	req.SetOperation(getOr(cfg, "method", "GET"))
	req.SetUrl(cfg["url"])
	req.SetKeepAlive(strings.EqualFold(getOr(cfg, "keepalive", "true"), "true"))
	req.SetProtocol(cfg["protocol"])
	req.HttpHeader = extractHeaders(cfg)
	return req
}

func getOr(cfg map[string]string, key, defaultVal string) string {
	if v, ok := cfg[key]; ok {
		return v
	}
	return defaultVal
}

func extractHeaders(cfg map[string]string) map[string]string {
	headers := make(map[string]string)
	for k, v := range cfg {
		if after, ok := strings.CutPrefix(k, "header."); ok {
			headers[after] = v
		}
	}
	return headers
}
