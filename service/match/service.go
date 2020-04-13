package match

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
	"nakama-golang/model"
)

type Service struct {
	Source       map[int64]string // mmr => id
	Players      chan string
	ReadyChan    map[string]chan *model.PlayerRealTime
	readyMutex   sync.Mutex
	Topic        string
	ctx          context.Context
	logger       runtime.Logger
	db           *sql.DB
	nk           runtime.NakamaModule
	defaultParam map[string]interface{}
	Match        chan *model.Match
}

func New(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, topic string) *Service {
	s := &Service{
		Source:       map[int64]string{},
		Players:      make(chan string, 1024*1024),
		ReadyChan:    map[string]chan *model.PlayerRealTime{},
		readyMutex:   sync.Mutex{},
		Topic:        topic,
		db:           db,
		nk:           nk,
		ctx:          ctx,
		logger:       logger,
		defaultParam: map[string]interface{}{},
		Match:        make(chan *model.Match, 1024),
	}
	go s.run()
	return s
}

func (s *Service) AddPlayer(id string) {
	s.Players <- id
}

func (s *Service) run() {
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case player := <-s.Players:
			// add player to mmrMap
			s.Source[time.Now().Unix()] = player
		case <-ticker.C:
			s.match()
		}
	}
}

func (s *Service) match() {
	player := []string{}
	count := 0
	for k, v := range s.Source {
		if count > 5 {
			break
		}
		s.logger.Info("match :%v=>%v", k, v)
		player = append(player, v)
		count++
	}
	if len(player) < 5 {
		return
	}
	matchId, err := s.nk.MatchCreate(s.ctx, s.Topic, s.defaultParam)
	if err != nil {
		s.logger.Error("match create err:%+v", err)
		return
	}
	info := map[string]interface{}{
		"players": player,
		"matchId": matchId,
	}
	for _, v := range player {
		if err := s.nk.NotificationSend(s.ctx, v, "match", info, 0, "", false); err != nil {
			s.logger.Error("match notify err:%+v", err)
			return
		}
	}
	count = 0
	for k, _ := range s.Source {
		count++
		if count > 5 {
			break
		}
		delete(s.Source, k)
	}
	pmap := map[string]string{}
	for _, v := range player {
		pmap[v] = ""
	}
	ma := &model.Match{
		MatchId: matchId,
		Players: pmap,
	}
	s.readyMutex.Lock()
	defer s.readyMutex.Unlock()
	s.ReadyChan[matchId] = make(chan *model.PlayerRealTime, 16)
	go s.ready(ma, s.ReadyChan[matchId])
}

func (s *Service) Start(mat *model.Match) {
	player:=[]string{}
	for id,_:=range mat.Players{
		player=append(player,id)
	}
	info := map[string]interface{}{
		"players": player,
		"matchId": mat.MatchId,
	}
	for v, _ := range mat.Players {
		if err := s.nk.NotificationSend(s.ctx, v, "match_start", info, 0, "", false); err != nil {
			s.logger.Error("match notify err:%+v", err)
			return
		}
	}
	s.Match <- mat
}

func (s *Service) Rejoin(mat *model.Match) {
	for p, _ := range mat.Players {
		// todo notify failed
		s.AddPlayer(p)
	}
}

func (s *Service) ready(mat *model.Match, ch chan *model.PlayerRealTime) {
	ticker := time.NewTicker(time.Second)
	timer := time.NewTimer(time.Second * 30)
	defer func() {
		s.readyMutex.Lock()
		defer s.readyMutex.Unlock()
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
		case <-timer.C:
			s.logger.Info("match :%v => failed", mat.MatchId)
			s.Rejoin(mat)
			return

		case player := <-ch:
			mat.Players[player.UserId] = player.SessionId
			// todo notify
		}
	}
}

func (s *Service) ReadyMatch(matchId string, player, sessionID string) {
	info := &model.PlayerRealTime{
		UserId:    player,
		SessionId: sessionID,
	}
	s.readyMutex.Lock()
	defer s.readyMutex.Unlock()
	if _, exist := s.ReadyChan[matchId]; exist {
		s.ReadyChan[matchId] <- info
	}
}
