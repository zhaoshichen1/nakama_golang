package game

import (
	"context"
	"database/sql"

	"github.com/heroiclabs/nakama-common/runtime"
	"nakama-golang/model"
)

type Service struct {
	Ctx context.Context
	Logger runtime.Logger
	Db *sql.DB
	Nk runtime.NakamaModule
}

func New(ctx context.Context, logger runtime.Logger, db *sql.DB,nk runtime.NakamaModule)*Service{
	s:=&Service{
		Ctx:ctx,
		Logger:logger,
		Db:db,
		Nk:nk,
	}
	return s
}

func (s *Service)Start(match *model.Match){
	s.initGame(match)
}

func (s *Service)initGame(match *model.Match){
	msgs:=[]*runtime.NotificationSend{}
	for i:=range match.Players{
		tmp:=&runtime.NotificationSend{
			UserID:     match.Players[i],
			Subject:    "match_init",
			Content:    nil,
			Code:       0,
			Sender:     "",
			Persistent: false,
		}
		sessionId,ok:=s.Ctx.Value("RUNTIME_CTX_SESSION_ID").(string)
		if !ok{
			return
		}
		if ok, err := s.Nk.StreamUserJoin(model.MatchStream, match.MatchId, "", "", match.Players[i], sessionId, false,false, "");err!=nil||!ok{

		}
		msgs=append(msgs,tmp)
	}
	if err:=s.Nk.NotificationsSend(s.Ctx,msgs);err!=nil{
		s.Logger.Error("initGame err:%+v",err)
		return
	}
}