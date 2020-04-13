package service

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/heroiclabs/nakama-common/api"
	"nakama-golang/model/event"
	"nakama-golang/protocol"
	"nakama-golang/util"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/pkg/errors"
)

func (s *Service) Hello(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, req protocol.Request) (v interface{}, err error) {
	userId, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	if !ok {
		// User ID not found in the context.
		return "", errors.New("userId not found!")
	}
	logger.Info("userId:%v", userId)

	s.Notify(ctx, logger, db, nk, userId, &protocol.NotifyHelloMsg{})
	return req, nil
}

func (s *Service) MatchJoin(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, req *protocol.ReqMatchJoin) (v interface{}, err error) {
	if err := nk.Event(ctx, &api.Event{
		Name: event.EventMatchJoin.String(),
		Properties: map[string]string{
			"topic": req.Topic,
			"aid":   util.ToString(req.Aid),
		},
		External: false,
	}); err != nil {
		return nil, err
	}
	return nil, nil
}

func (s *Service) MatchReady(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, req *protocol.ReqMatchReady) (v interface{}, err error) {
	if err := nk.Event(ctx, &api.Event{
		Name: event.EventMatchReady.String(),
		Properties: map[string]string{
			"match_id": req.MatchId,
			"aid":      util.ToString(req.MatchId),
		},
		External: false,
	}); err != nil {
		return nil, err
	}
	return nil, nil
}

func (s *Service) GameTick(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, req *protocol.ReqGameTick) (v interface{}, err error) {
	jstr, _ := json.Marshal(req.Frame)
	if err := nk.Event(ctx, &api.Event{
		Name: event.EventGameRun.String(),
		Properties: map[string]string{
			"data": string(jstr),
		},
		External: false,
	}); err != nil {
		return nil, err
	}
	return nil, nil
}
