package model

const (
	GameStream      = 10
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
