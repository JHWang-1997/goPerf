package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func FileExist(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func ParseProperties(path string) (*map[string]string, error) {
	open, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer open.Close()
	res := make(map[string]string)
	scanner := bufio.NewScanner(open)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			// 跳过注释
			continue
		}
		if !strings.ContainsRune(line, '=') {
			return nil, fmt.Errorf("Invalited item: %s", line)
		}
		split := strings.Split(line, "=")
		if len(split) != 2 {
			return nil, fmt.Errorf("Invalited item: %s", line)
		}
		key := split[0]
		val := split[1]
		res[key] = val
	}
	return &res, nil
}
