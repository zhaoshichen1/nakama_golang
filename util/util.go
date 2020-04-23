package util

import (
	"strconv"
)

func ToInt64(str string) int64 {
	v, _ := strconv.ParseInt(str, 10, 64)
	return v
}
