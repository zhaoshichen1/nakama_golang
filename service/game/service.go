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
const(
	Init = Status("init")
	Ready = Status("ready")
	Running = Status("running")
	Finish = Status("finish")
)

type Group struct {
	sync.Mutex
	group map[string]*Service
	Ctx    context.Context
	Logger runtime.Logger
	Db     *sql.DB
	Nk     runtime.NakamaModule
	closeChan chan string
}

func NewGroup(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule)*Group{
	g:=&Group{
		group: map[string]*Service{},
		Mutex:sync.Mutex{},
		Ctx:    ctx,
		Logger: logger,
		Db:     db,
		Nk:     nk,
		closeChan:make(chan string,1024*64),
	}
	return g
}
func (g *Group)Start(match *model.Match){
	g.Lock()
	defer g.Unlock()
	s:=New(g.Ctx,g.Logger,g.Db,g.Nk,match, func() {
		g.closeChan<-match.MatchId
	})
	if _,exist:=g.group[match.MatchId];!exist{
		g.group[match.MatchId]=s
	}
	s.Start(match)
}

func (g *Group)Tick(){
	for{
		select {
		case matchId,ok:=<-g.closeChan:
			if !ok {
				return
			}
			delete(g.group, matchId)
		}
	}
}

func (g *Group)Run(msg *model.GameMsg){
	g.Lock()
	defer g.Unlock()
	if _,exist:=g.group[msg.MatchId];exist {
		g.group[msg.MatchId].Run(msg)
	}
}

type Service struct {
	Aid int64
	Ctx    context.Context
	Logger runtime.Logger
	Db     *sql.DB
	Nk     runtime.NakamaModule
	match *model.Match
	closed bool
	closeFunc func()
	gut chan *model.GameMsg
	curTick int64
	PlayerTick map[string]map[int64]*model.GamePlayFrame
	TimeTick map[int64]map[string]*model.GamePlayFrame
	status Status
}

func New(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule,match *model.Match,closeFunc func()) *Service {
	s := &Service{
		Aid:match.Aid,
		Ctx:    ctx,
		Logger: logger,
		Db:     db,
		Nk:     nk,
		gut:make(chan *model.GameMsg,1024*1024),
		PlayerTick: map[string]map[int64]*model.GamePlayFrame{},
		TimeTick: map[int64]map[string]*model.GamePlayFrame{},
		status:Init,
		match:match,
		closeFunc:closeFunc,
	}
	return s
}

func (s *Service)close(){
	// todo
	s.closed=true
	close(s.gut)
	s.closeFunc()
}
func (s *Service)set(st Status){
	s.status=st
}

func (s *Service) Start(match *model.Match) {
	for{
		switch s.status {
		case Init:
			s.initGame()
			s.set(Ready)
		case Ready:

		case Running:
			s.startGame()
			s.run()
		case Finish:
			s.finishGame()
			break
		}
	}
	s.close()
}

func (s *Service)Run(msg *model.GameMsg){
	if !s.closed {
		s.gut <- msg
	}
}

func (s *Service) initGame() {
	msgs := []*runtime.NotificationSend{}
	for id, sea := range s.match.Players {
		tmp := &runtime.NotificationSend{
			UserID:     id,
			Subject:    "game_init",
			Content:    nil,
			Code:       0,
			Sender:     "",
			Persistent: false,
		}
		if ok, err := s.Nk.StreamUserJoin(model.GameStream, s.match.MatchId, "", "", id, sea, false, false, ""); err != nil || !ok {
			s.Logger.Error("join failed err:%+v",err)
			return
		}
		msgs = append(msgs, tmp)
	}
	if err := s.Nk.NotificationsSend(s.Ctx, msgs); err != nil {
		s.Logger.Error("initGame err:%+v", err)
		return
	}
}

func (s *Service)startGame(){
	msgs := []*runtime.NotificationSend{}
	for id, _ := range s.match.Players {
		tmp := &runtime.NotificationSend{
			UserID:     id,
			Subject:    "game_start",
			Content:    nil,
			Code:       0,
			Sender:     "",
			Persistent: false,
		}
		msgs = append(msgs, tmp)
	}
	if err := s.Nk.NotificationsSend(s.Ctx, msgs); err != nil {
		s.Logger.Error("start_game err:%+v", err)
		return
	}
}

func (s *Service)finishGame(){
	msgs := []*runtime.NotificationSend{}
	for id, _ := range s.match.Players {
		tmp := &runtime.NotificationSend{
			UserID:     id,
			Subject:    "game_finish",
			Content:    nil,
			Code:       0,
			Sender:     "",
			Persistent: false,
		}
		msgs = append(msgs, tmp)
	}
	if err := s.Nk.NotificationsSend(s.Ctx, msgs); err != nil {
		s.Logger.Error("finish_game err:%+v", err)
		return
	}
}

func (s *Service)isFinish()bool{
	return s.curTick>=100
}

func (s *Service)finish(){
	if err:=s.Nk.StreamClose(model.GameStream,s.match.MatchId,"","",);err!=nil{
		s.Logger.Error("finish StreamClose err:%+v",err)
	}
	s.set(Finish)
}

func (s *Service)run(){
	// 50ms 下发数据
	ticker:=time.NewTicker(time.Millisecond*50)
	s.curTick=0
	for{
		select {
		case <-ticker.C:
			s.stream()
			s.curTick++
			if s.isFinish(){
				s.finish()
				return
			}
		case msg,closed:=<-s.gut:
			if !closed{
				return
			}
			s.process(msg)
		}
	}
}

func (s *Service)stream(){
	cur:=s.TimeTick[s.curTick]
	jstr,err:=json.Marshal(cur)
	if err!=nil{
		s.Logger.Error("stream Marshal err:%+v",err)
		return
	}
	if err:=s.Nk.StreamSend(model.GameStream,s.match.MatchId,"","",string(jstr),nil, true);err!=nil{
		s.Logger.Error("stream StreamSend err:%+v",err)
	}
}

func (s *Service)process(msg *model.GameMsg){
	// todo point
	msg.Point=rand.Int63()%50+50
	s.PlayerTick[msg.UserId][s.curTick]=msg.Data
	s.TimeTick[s.curTick][msg.UserId]=msg.Data
}