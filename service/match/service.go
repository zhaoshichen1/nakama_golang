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
	Source       []string // 用户池
	players      chan string
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
		Source:       make([]string, 0),
		players:      make(chan string, 1024*1024),
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
	s.players <- id
}

func (s *Service) run() {
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case player := <-s.players: // add player to mmrMap
			s.Source = append(s.Source, player)
		case <-ticker.C: // every second try to match
			s.match()
		}
	}
}

func (s *Service) match() {
	if len(s.Source) < model.MathMinPlayers {
		return
	}
	players := s.Source[:model.MathMinPlayers] // pick minimum users to play
	s.Source = s.Source[model.MathMinPlayers:]
	matchId, err := s.nk.MatchCreate(s.ctx, s.Topic, s.defaultParam)
	if err != nil {
		s.logger.Error("match create err:%+v", err)
		return
	}
	info := map[string]interface{}{
		"players": players,
		"matchId": matchId,
	}
	for _, v := range players {
		if err := s.nk.NotificationSend(s.ctx, v, "match", info, 0, "", false); err != nil {
			s.logger.Error("match notify err:%+v", err)
			return
		}
	}
	pmap := map[string]string{}
	for _, v := range players {
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
	info := map[string]interface{}{
		"players": mat.Players,
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
	for p := range mat.Players {
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
