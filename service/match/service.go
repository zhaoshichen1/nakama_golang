package match

import (
	"context"
	"database/sql"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
	"nakama-golang/model"
)

type Service struct {
	Source       map[int64]string
	Players      chan string
	Ready        map[string]map[string]bool
	ReadyChan    map[string]chan string
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
		Ready:        map[string]map[string]bool{},
		ReadyChan:    map[string]chan string{},
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
	pmap := map[string]bool{}
	for _, v := range player {
		pmap[v] = false
	}
	s.Ready[matchId] = pmap
	s.ReadyChan[matchId] = make(chan string, 16)
	go s.ready(matchId)
}

func (s *Service) Start(matchId string) {
	player := s.Ready[matchId]
	info := map[string]interface{}{
		"players": player,
		"matchId": matchId,
	}
	for v, _ := range player {
		if err := s.nk.NotificationSend(s.ctx, v, "match_start", info, 0, "", false); err != nil {
			s.logger.Error("match notify err:%+v", err)
			return
		}
	}
}

func (s *Service) Rejoin(matchId string) {
	delete(s.ReadyChan, matchId)
	player := s.Ready[matchId]
	delete(s.Ready, matchId)
	for p, _ := range player {
		// todo notify failed
		s.AddPlayer(p)
	}
}

func (s *Service) ready(matchId string) {
	ticker := time.NewTicker(time.Second)
	timer := time.NewTimer(time.Second * 30)
	for {
		select {
		case <-ticker.C:
			for k, v := range s.Ready[matchId] {
				s.logger.Info("match :%v => player:%v ready status:%v", matchId, k, v)
				if !v {
					continue
				}
				s.Start(matchId)
				return
			}
		case <-timer.C:
			s.logger.Info("match :%v => failed", matchId)
			s.Rejoin(matchId)
			return

		case player := <-s.ReadyChan[matchId]:
			s.Ready[matchId][player] = true
		}
	}
}

func (s *Service) ReadyMatch(matchId string, player string) {
	if _, exist := s.ReadyChan[matchId]; exist {
		s.ReadyChan[matchId] <- player
	}
}
