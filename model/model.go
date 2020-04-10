package model

import "github.com/heroiclabs/nakama-common/runtime"

type NotifyType string

const (
	NotifyHello NotifyType = "NotifyHello"
)

type Request interface {
	Meta() string
}

// content map[string]interface{}, code int, sender string, persistent bool
type Notify interface {
	Type() NotifyType
	Subject() string
	Data() map[string]interface{}
	Code() int
	Sender() string
	Persistent() bool
}

type NotifyCommon struct {
	runtime.NotificationSend
}

func (v *NotifyCommon)Subject()string{
	return v.NotificationSend.Subject
}

func (v *NotifyCommon) Data() map[string]interface{} {
	return v.NotificationSend.Content
}

func (v *NotifyCommon) Code() int {
	return v.NotificationSend.Code
}

func (v *NotifyCommon) Sender() string {
	return v.NotificationSend.Sender
}

func (v *NotifyCommon) Persistent() bool {
	return v.NotificationSend.Persistent
}
