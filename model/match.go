package model

const (
	StreamGameData  = 10
	StreamGameMsg   = 11
	ConfirmDeadline = 30 // 确认时间30s
)

type Match struct {
	Aid     int64
	MatchId string
	Players map[string]string
}

type PlayerRealTime struct {
	UserId    string
	SessionId string
}
