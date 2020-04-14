package protocol

import "nakama-golang/model"

type ReqMatchJoin struct {
	Aid   int64  `json:"aid"`
	Topic string `json:"topic"`
}

func (v *ReqMatchJoin) Meta() string {
	return ""
}

type ResMatchJoin struct {
	Topic   string `json:"topic"`
	MatchId string `json:"match_id"`
}

type ReqMatchReady struct {
	Topic   string `json:"topic"`
	MatchId string `json:"match_id"`
	Aid     int64  `json:"aid"`
}

type ReqGameTick struct {
	CurTick int64
	Frame   *model.GamePlayFrame
}

type ReqGameReady struct {
}
