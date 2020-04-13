package protocol

import (
	"encoding/json"
)

type ReqHello struct {
	Hello string `json:"hello"`
}

func (v *ReqHello) Meta() string {
	jstr, _ := json.Marshal(v)
	return string(jstr)
}

type NotifyHelloMsg struct {
	NotifyCommon
}

func (v *NotifyHelloMsg) Type() NotifyType {
	return NotifyHello
}

func (v *NotifyHelloMsg) Data() map[string]interface{} {
	v.Content["notify_type"] = NotifyHello
	return v.NotificationSend.Content
}
