package model

const (
	GameStream = 10
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
