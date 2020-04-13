package match

import (
	"context"
	"sync"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"

	"nakama-golang/model"
)

// Manager 为所有aid进行匹配
type Manager struct {
	// match structures
	Groups   map[int64]*Group // key=aid,value=match_group
	mutex    sync.RWMutex
	NewMatch chan *model.Match // 新匹配成功

	// dependency objects
	ctx    context.Context
	logger runtime.Logger
	nk     runtime.NakamaModule
}

func NewMatchManager() *Manager {
	return &Manager{
		Groups:   make(map[int64]*Group),
		NewMatch: make(chan *model.Match, 1024),
	}
}

// 新加入玩家
func (this *Manager) NewPlayer(aid int64, userIDS []string) {

	this.mutex.Lock()
	defer this.mutex.Unlock()
	if _, ok := this.Groups[aid]; !ok { // 创建新的match group
		this.Groups[aid] = &Group{}
	}

	this.Groups[aid].AddPlayer(userIDS) // 加入match group的匹配队列
}

// 开启定期查找是否有符合要求的玩家组合可以匹配
func (this *Manager) Match() {
	ticker := time.NewTicker(time.Second)

	try := func() {
		this.mutex.Lock()
		defer this.mutex.Unlock()
		for aid, v := range this.Groups { // 每个aid的分组都进行轮询匹配
			playerGroups := v.Match()
			if len(playerGroups) == 0 {
				continue
			}
			for _, players := range playerGroups {
				newMatch, err := this.nk.MatchCreate(context.Background(), "", defaultParam)
				if err != nil {
					this.logger.Printf("match create uids %v, err %v", players, err)
					this.NewPlayer(aid, players) // 匹配失败，重新加入等待队列中
					continue
				}
				// tmp := make(map[string]string)
				// for _, v := range players {
				// 	tmp[v] = ""
				// }
				// this.NewMatch <- &model.Match{ // 通知game service新的匹配诞生
				// 	Aid:     aid,
				// 	MatchId: newMatch,
				// 	Players: tmp,
				// }
			}
		}
	}

	for {
		select {
		case <-ticker.C:
			try()
		}
	}
}

func (this *Manager) Start(mat *model.Match) {
	info := map[string]interface{}{
		"match_id": mat.MatchId,
		"deadline": model.ConfirmDeadline,
	}
	for userID := range mat.Players {
		if err := this.nk.NotificationSend(context.Background(), userID, "match_start", info, 0, "", true); err != nil {
			this.logger.Error("match notify err:%+v", err)
			return
		}
	}
}

func (s *Match) ready(mat *model.Match, ch chan *model.PlayerRealTime) {
	ticker := time.NewTicker(time.Second)
	timer := time.NewTimer(time.Second * model.ConfirmDeadline)
	defer func() {
		s.mutex.Lock()
		defer s.mutex.Unlock()
		delete(s.ReadyChan, mat.MatchId)
	}()
	for {
		select {
		case <-ticker.C:
			for k, v := range mat.Players {
				s.logger.Info("match :%v => player:%v ready status:%v", mat.MatchId, k, v)
				if v == "" {
					continue
				}
				s.Start(mat)
				return
			}
		case <-timer.C: // todo 通知匹配超时
			s.logger.Info("match :%v => failed", mat.MatchId)
			return
		case player := <-ch:
			mat.Players[player.UserId] = player.SessionId
			// todo notify
		}
	}
}

func (s *Match) ReadyMatch(matchId string, player, sessionID string) {
	info := &model.PlayerRealTime{
		UserId:    player,
		SessionId: sessionID,
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if _, exist := s.ReadyChan[matchId]; exist {
		s.ReadyChan[matchId] <- info
	}
}
