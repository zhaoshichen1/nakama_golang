package model

const (
	MatchStream    = 10
	MathMinPlayers = 2
)

type Match struct {
	MatchId string
	Players map[string]string
}

type PlayerRealTime struct {
	UserId    string
	SessionId string
}
