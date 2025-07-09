package utils

import (
	"fmt"
	"time"
)

func Format(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

func NowNano() uint64 {
	return uint64(time.Now().UnixNano())
}
