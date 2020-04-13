package util

import (
	"fmt"
	"strconv"
)

func ToString(v interface{})string{
	return fmt.Sprintf("%v",v)
}
func ToInt64(str string)int64{
	v,_:=strconv.ParseInt(str,10,64)
	return v
}
