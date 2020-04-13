package model

type GameMsg struct {
	UserId    string
	SessionId string
	MatchId   string
	Data      *GamePlayFrame
	Point     int64
}

type GamePlayFrame struct {
}
