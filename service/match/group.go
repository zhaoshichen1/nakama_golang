package match

import (
	"sync"

	"nakama-golang/model"
)

var defaultParam = make(map[string]interface{})

// Group 为同一个aid下的玩家进行多组匹配
type Group struct {
	Players     []string
	playerMutex sync.Mutex

	Matches  map[string]*model.Match
	matMutex sync.Mutex
}

func (g *Group) AddPlayer(UserIds []string) {
	g.playerMutex.Lock()
	defer g.playerMutex.Unlock()
	g.Players = append(g.Players, UserIds...)
}

func (g *Group) Match() (res [][]string) {
	if len(g.Players) < model.MatchMinPlayer { // 不足匹配人数
		return
	}
	var tmp []string
	for _, v := range g.Players {
		tmp = append(tmp, v)
		if len(tmp) == model.MatchMinPlayer {
			res = append(res, append([]string{}, tmp...))
			tmp = []string{}
		}
	}
	g.playerMutex.Lock()
	defer g.playerMutex.Unlock()
	g.Players = g.Players[len(res)*model.MatchMinPlayer:] // 截断已经匹配成功的players
	return
}
