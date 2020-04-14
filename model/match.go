package model

const (
	StreamGameData  = 10
	StreamGameMsg   = 11
	ConfirmDeadline = 30 // 确认时间30s
	MatchMinPlayer  = 2
)

type Match struct {
	Aid     int64
	MatchId string
	Players map[string]string // key=userID, value=sessionID
	Chan    chan *PlayerRealTime
}

type PlayerRealTime struct {
	UserId    string
	SessionId string
}
