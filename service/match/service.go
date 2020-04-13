package match

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
	"nakama-golang/model"
)

type Group struct {
	group map[int64]*Service
	Match chan *model.Match
	sync.Mutex
}

func NewGroup() *Group {
	g := &Group{
		group: map[int64]*Service{},
		Match: make(chan *model.Match, 1024*64),
		Mutex: sync.Mutex{},
	}
	return g
}
func (g *Group) Add(s *Service) {
	g.Lock()
	defer g.Unlock()
	if _, exist := g.group[s.Aid]; !exist {
		g.group[s.Aid] = s
	}
	// chan转发
	go func() {
		for {
			if v, ok := <-s.Match; ok {
				g.Match <- v
			}
			return
		}
	}()
}
func (g *Group) AddPlayer(Aid int64, UserId string) {
	g.Lock()
	defer g.Unlock()
	ser, exist := g.group[Aid]
	if exist {
		ser.AddPlayer(UserId)
	}
}

func (g *Group) ReadyMatch(Aid int64, matchId string, player, sessionID string) {
	g.Lock()
	defer g.Unlock()
	ser, exist := g.group[Aid]
	if exist {
		ser.ReadyMatch(matchId, player, sessionID)
	}
}

type Service struct {
	Aid          int64
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

func New(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, aid int64, topic string) *Service {
	s := &Service{
		Aid:          aid,
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
		if count > model.MatchMinPlayer {
			break
		}
		s.logger.Info("match :%v=>%v", k, v)
		player = append(player, v)
		count++
	}
	if len(player) < model.MatchMinPlayer {
		return
	}
	matchId, err := s.nk.MatchCreate(s.ctx, s.Topic, s.defaultParam)
	if err != nil {
		s.logger.Error("match create err:%+v", err)
		return
	}
	info := map[string]interface{}{
		"players":  player,
		"match_id": matchId,
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
		if count > model.MatchMinPlayer {
			break
		}
		delete(s.Source, k)
	}
	pmap := map[string]string{}
	for _, v := range player {
		pmap[v] = ""
	}
	ma := &model.Match{
		Aid:     s.Aid,
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
		"match_id": mat.MatchId,
		"deadline": model.ConfirmDeadline,
	}
	for userID := range mat.Players {
		// todo fail-over
		if err := s.nk.NotificationSend(s.ctx, userID, "match_start", info, 0, "", true); err != nil {
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
	timer := time.NewTimer(time.Second * model.ConfirmDeadline)
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
