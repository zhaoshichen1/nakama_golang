package model

const (
	GameStream      = 10
	ConfirmDeadline = 30 // 确认时间30s
	MatchMinPlayer  = 2
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
