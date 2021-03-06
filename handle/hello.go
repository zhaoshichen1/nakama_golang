package handle

import (
	"nakama-golang/fantasy"
	"nakama-golang/protocol"
)

func gameReady(t *fantasy.Tifa) {
	req := &protocol.ReqGameReady{}
	if err := t.Bind(req); err != nil {
		t.Abort()
		return
	}
	t.Json(ser.GameReady(t.Ctx, t.Logger, t.Db, t.Nk, req))
}

func gameTick(t *fantasy.Tifa) {
	req := &protocol.ReqGameTick{}
	if err := t.Bind(req); err != nil {
		t.Abort()
		return
	}
	t.Json(ser.GameTick(t.Ctx, t.Logger, t.Db, t.Nk, req))
}
