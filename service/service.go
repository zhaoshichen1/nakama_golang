package service

import (
	"context"
	"database/sql"

	"github.com/heroiclabs/nakama-common/runtime"
	"nakama-golang/protocol"
)

type Service struct {

}

func New() (s *Service) {
	s = &Service{}
	if err := s.init(); err != nil {
		panic(err)
	}
	return s
}

func (s *Service) init() error {

	return nil
}

func (s *Service) Close() {

}

func (s *Service) Notify(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, userId string, msg protocol.Notify) {
	if err := nk.NotificationSend(ctx, userId, msg.Subject(), msg.Data(), msg.Code(), msg.Sender(), msg.Persistent()); err != nil {
		logger.Error("NotificationSend err:%v", err)
	}
}
