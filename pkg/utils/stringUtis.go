package utils

import (
	"fmt"
	"strconv"
)

func ParseInt(val string, valName string) int {
	num, err := strconv.Atoi(val)
	if err != nil {
		panic(fmt.Errorf("%s should be Integer!", valName))
	}
	return num
}
