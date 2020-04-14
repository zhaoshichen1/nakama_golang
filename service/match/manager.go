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
	Groups       map[int64]*Group // key=aid,value=match_group
	mutex        sync.RWMutex
	NewMatch     chan *model.Match // 新匹配成功
	ConfirmMatch chan *model.PlayerRealTime

	// dependency objects
	ctx    context.Context
	logger runtime.Logger
	nk     runtime.NakamaModule
}

func NewMatchManager(ctx context.Context) *Manager {
	return &Manager{
		ctx:          ctx,
		Groups:       make(map[int64]*Group),
		NewMatch:     make(chan *model.Match, 10240),
		ConfirmMatch: make(chan *model.PlayerRealTime, 10240),
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
				newMatch, err := this.nk.MatchCreate(this.ctx, "", defaultParam)
				if err != nil {
					this.logger.Printf("match create uids %v, err %v", players, err)
					this.NewPlayer(aid, players) // 匹配失败，重新加入等待队列中
					continue
				}
				tmp := make(map[string]string)
				for _, v := range players { // 等待用户确认上报sessionID
					tmp[v] = ""
				}

				this.Groups[aid].matMutex.Lock() // 进入等待确认流程
				match := &model.Match{
					Aid:     aid,
					MatchId: newMatch,
					Players: tmp,
					Chan:    make(chan *model.PlayerRealTime, 1024),
				}
				this.Groups[aid].Matches[newMatch] = match
				this.Groups[aid].matMutex.Unlock()
				go this.WaitConfirm(match, match.Chan)
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
		if err := this.nk.NotificationSend(this.ctx, userID, "match_start", info, 0, "", true); err != nil {
			this.logger.Error("match notify err:%+v", err)
			return
		}
	}
	this.NewMatch <- mat // 通知game那边有新确认的匹配信息，准备开始游戏
}

// 等待确认
func (this *Manager) WaitConfirm(mat *model.Match, ch chan *model.PlayerRealTime) {
	defer func() { // 删除匹配信息
		this.Groups[mat.Aid].matMutex.Lock()
		delete(this.Groups[mat.Aid].Matches, mat.MatchId)
		this.Groups[mat.Aid].matMutex.Unlock()
	}()
	ticker := time.NewTicker(time.Second)
	timer := time.NewTimer(time.Second * model.ConfirmDeadline)
	for {
		select {
		case <-ticker.C:
			if finished := func() bool {
				for k, v := range mat.Players { // 等待用户请求上报session ID
					this.logger.Info("match :%v => player:%v ready status:%v", mat.MatchId, k, v)
					if v == "" {
						return false
					}
				}
				this.Start(mat) // start放到外层
				return true
			}(); finished { // 所有人确认完毕
				return
			}
		case <-timer.C: // todo 通知匹配超时
			this.logger.Info("match :%v => failed", mat.MatchId)
			return
		case player := <-ch:
			mat.Players[player.UserId] = player.SessionId
		}
	}
}

func (this *Manager) ReadyMatch(aid int64, matchId string, player, sessionID string) {
	info := &model.PlayerRealTime{
		UserId:    player,
		SessionId: sessionID,
	}
	if _, ok := this.Groups[aid]; !ok {
		this.logger.Printf("Aid Group not found, aid %d", aid)
		return
	}
	this.Groups[aid].matMutex.Lock() // 进入等待确认流程
	if _, ok := this.Groups[aid].Matches[matchId]; !ok {
		this.logger.Printf("Aid %d MatchID %s", aid, matchId)
		return
	}
	this.Groups[aid].Matches[matchId].Chan <- info // 计入确认信息
	this.Groups[aid].matMutex.Unlock()
}
