package game

import (
	"context"
	"database/sql"
	"encoding/json"
	"math/rand"
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
type Service struct {
	Ctx    context.Context
	Logger runtime.Logger
	Db     *sql.DB
	Nk     runtime.NakamaModule
	match *model.Match
	gut chan *model.GameMsg
	curTick int64
	PlayerTick map[string]map[int64]*model.GamePlayFrame
	TimeTick map[int64]map[string]*model.GamePlayFrame
	status Status
}

func New(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule) *Service {
	s := &Service{
		Ctx:    ctx,
		Logger: logger,
		Db:     db,
		Nk:     nk,
		gut:make(chan *model.GameMsg,1024*1024),
		PlayerTick: map[string]map[int64]*model.GamePlayFrame{},
		TimeTick: map[int64]map[string]*model.GamePlayFrame{},
		status:Init,
	}
	return s
}

func (s *Service)set(st Status){
	s.status=st
}

func (s *Service) Start(match *model.Match) {
	s.match=match
	for{
		switch s.status {
		case Init:
			s.initGame()
			s.set(Ready)
		case Ready:

		case Running:
			s.startGame()
			go s.run()
		case Finish:
			s.finishGame()
			return
		}
	}
}

func (s *Service)Run(msg *model.GameMsg){
	s.gut<-msg
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
			// todo send to client
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