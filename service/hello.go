package service

import (
	"context"
	"database/sql"
	"nakama-golang/model"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/pkg/errors"
)

func (s *Service) Hello(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, req model.Request) (v interface{}, err error) {
	userId, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	if !ok {
		// User ID not found in the context.
		return "", errors.New("userId not found!")
	}
	logger.Info("userId:%v", userId)

	s.Notify(ctx, logger, db, nk, userId, &model.NotifyHelloMsg{})
	return req, nil
}
