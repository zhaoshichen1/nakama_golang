package game

import (
	"context"
	"database/sql"
	"encoding/json"
	"math/rand"
	"sync"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
	"nakama-golang/model"
)

type Status string

const (
	Init    = Status("game_init")
	Ready   = Status("game_ready")
	Start   = Status("game_start")
	Running = Status("game_running")
	Finish  = Status("game_finish")
)

func (s Status) String() string {
	return string(s)
}

type Group struct {
	sync.Mutex
	group     map[string]*Service
	Ctx       context.Context
	Logger    runtime.Logger
	Db        *sql.DB
	Nk        runtime.NakamaModule
	closeChan chan string
}

func NewGroup(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule) *Group {
	g := &Group{
		group:     map[string]*Service{},
		Mutex:     sync.Mutex{},
		Ctx:       ctx,
		Logger:    logger,
		Db:        db,
		Nk:        nk,
		closeChan: make(chan string, 1024*64),
	}
	return g
}

func (g *Group) Start(match *model.Match) {
	g.Lock()
	defer g.Unlock()
	s := New(g.Ctx, g.Logger, g.Db, g.Nk, match, func() {
		g.closeChan <- match.MatchId
	})
	if _, exist := g.group[match.MatchId]; !exist {
		g.group[match.MatchId] = s
	}
	s.Start(match)
}

func (g *Group) Tick() {
	for {
		select {
		case matchId, ok := <-g.closeChan:
			if !ok {
				return
			}
			delete(g.group, matchId)
		}
	}
}

func (g *Group) Run(msg *model.GameMsg) {
	g.Lock()
	defer g.Unlock()
	if _, exist := g.group[msg.MatchId]; exist {
		g.group[msg.MatchId].Run(msg)
	}
}

type Service struct {
	Aid         int64
	Ctx         context.Context
	Logger      runtime.Logger
	Db          *sql.DB
	Nk          runtime.NakamaModule
	match       *model.Match
	closed      bool
	closeFunc   func()
	gut         chan *model.GameMsg
	curTick     int64
	startTick   int64
	startTimer  *time.Timer
	PlayerFrame map[string]map[int64]*model.GamePlayFrame
	TimeFrame   map[int64]map[string]*model.GamePlayFrame
	status      Status
}

func New(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, match *model.Match, closeFunc func()) *Service {
	s := &Service{
		Aid:         match.Aid,
		Ctx:         ctx,
		Logger:      logger,
		Db:          db,
		Nk:          nk,
		gut:         make(chan *model.GameMsg, 1024*1024),
		PlayerFrame: map[string]map[int64]*model.GamePlayFrame{},
		TimeFrame:   map[int64]map[string]*model.GamePlayFrame{},
		status:      Init,
		match:       match,
		closeFunc:   closeFunc,
	}
	return s
}

func (s *Service) close() {
	// todo
	s.closed = true
	close(s.gut)
	s.closeFunc()
}
func (s *Service) set(st Status) {
	s.status = st
}

func (s *Service) Start(match *model.Match) {
	for {
		switch s.status {
		case Init:
			s.initGame()
		case Ready:
			s.readyGame()
		case Start:
			s.startGame()
		case Running:
			s.runGame()
		case Finish:
			s.finishGame()
			break
		}
	}
	s.close()
}

func (s *Service) Run(msg *model.GameMsg) {
	if !s.closed {
		s.gut <- msg
	}
}

func (s *Service) initGame() {
	for id, sea := range s.match.Players {
		if ok, err := s.Nk.StreamUserJoin(model.StreamGameData, s.match.MatchId, "", "", id, sea, false, false, ""); err != nil || !ok {
			s.Logger.Error("join failed err:%+v", err)
			return
		}
	}
	s.broadcast(Init.String(), nil)
	s.set(Ready)
}

func (s *Service) startGame() {
	s.startTick = time.Now().Add(time.Second * 5).Unix()
	s.startTimer = time.NewTimer(time.Second * 5)
	msg := &struct {
		StartTick int64
	}{
		StartTick: s.startTick,
	}
	s.broadcast(Start.String(), msg)
}

func (s *Service) readyGame() {
	// 超时
	timer := time.NewTimer(time.Minute)
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-timer.C:
			// todo failed notify
			s.set(Finish)
			return
		case <-ticker.C:
			if len(s.PlayerFrame) == len(s.match.Players) {
				// todo success notify
				s.set(Start)
				return
			}
		case msg, closed := <-s.gut:
			if !closed {
				return
			}
			s.ready(msg)
		}
	}
}

func (s *Service) broadcast(subject string, content interface{}) {
	jstr, _ := json.Marshal(content)
	if err := s.Nk.StreamSend(model.StreamGameMsg, subject, "", "", string(jstr), nil, true); err != nil {
		s.Logger.Warn("broadcast %s %+v failed err:%+v", subject, content, err)
	}
}

func (s *Service) ready(msg *model.GameMsg) {
	s.PlayerFrame[msg.UserId] = map[int64]*model.GamePlayFrame{}
	// todo notify
	s.broadcast(Ready.String(), map[string]interface{}{
		"player": msg.UserId,
		"status": s.status,
	})
}

func (s *Service) runGame() {
	select {
	case <-s.startTimer.C:
		s.broadcast(Running.String(), nil)
	}
	s.run()
}

func (s *Service) finishGame() {
	s.broadcast(Finish.String(), nil)
}

func (s *Service) isFinish() bool {
	return s.curTick >= 100
}

func (s *Service) finish() {
	if err := s.Nk.StreamClose(model.StreamGameData, s.match.MatchId, "", ""); err != nil {
		s.Logger.Error("finish StreamClose err:%+v", err)
	}
	s.set(Finish)
}

func (s *Service) run() {
	// 50ms 下发数据
	ticker := time.NewTicker(time.Millisecond * 50)
	s.curTick = 0
	for {
		select {
		case <-ticker.C:
			s.stream()
			s.curTick++
			if s.isFinish() {
				s.finish()
				return
			}
		case msg, closed := <-s.gut:
			if !closed {
				return
			}
			s.process(msg)
		}
	}
}

func (s *Service) stream() {
	cur := s.TimeFrame[s.curTick]
	jstr, err := json.Marshal(cur)
	if err != nil {
		s.Logger.Error("stream Marshal err:%+v", err)
		return
	}
	if err := s.Nk.StreamSend(model.StreamGameData, s.match.MatchId, "", "", string(jstr), nil, true); err != nil {
		s.Logger.Error("stream StreamSend err:%+v", err)
	}
}

func (s *Service) process(msg *model.GameMsg) {
	playerMp, exist := s.PlayerFrame[msg.UserId]
	if !exist {
		return
	}
	playerMp[s.curTick] = msg.Data
	// todo point
	msg.Point = rand.Int63()%50 + 50
	if _, exist := s.TimeFrame[s.curTick]; !exist {
		s.TimeFrame[s.curTick] = map[string]*model.GamePlayFrame{}
	}
	s.TimeFrame[s.curTick][msg.UserId] = msg.Data
}
