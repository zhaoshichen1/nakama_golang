package service

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/heroiclabs/nakama-common/api"

	"github.com/heroiclabs/nakama-common/runtime"

	"nakama-golang/model/event"
	"nakama-golang/protocol"
)

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

func (s *Service) GameReady(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, req *protocol.ReqGameReady) (v interface{}, err error) {
	userId, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	if !ok {
		// User ID not found in the context.
		return
	}
	sessionId, ok := ctx.Value(runtime.RUNTIME_CTX_SESSION_ID).(string)
	if !ok {
		// If session ID is not found, RPC was not called over a connected socket.
		return
	}
	matchId, ok := ctx.Value(runtime.RUNTIME_CTX_MATCH_ID).(string)
	if !ok {
		return
	}
	if err := nk.Event(ctx, &api.Event{
		Name: event.EventGameReady.String(),
		Properties: map[string]string{
			"user_id":    userId,
			"session_id": sessionId,
			"match_id":   matchId,
		},
		External: false,
	}); err != nil {
		return nil, err
	}
	return nil, nil
}
