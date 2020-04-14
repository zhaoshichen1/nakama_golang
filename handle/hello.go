package handle

import (
	"nakama-golang/fantasy"
	"nakama-golang/protocol"
)

func helloHandle(t *fantasy.Tifa) {
	req := &protocol.ReqHello{}
	if err := t.Bind(req); err != nil {
		t.Abort()
		return
	}
	t.Json(ser.Hello(t.Ctx, t.Logger, t.Db, t.Nk, req))
}

func matchHandle(t *fantasy.Tifa) {
	req := &protocol.ReqMatchJoin{}
	if err := t.Bind(req); err != nil {
		t.Abort()
		return
	}
	t.Json(ser.MatchJoin(t.Ctx, t.Logger, t.Db, t.Nk, req))
}

func matchReady(t *fantasy.Tifa) {
	req := &protocol.ReqMatchReady{}
	if err := t.Bind(req); err != nil {
		t.Abort()
		return
	}
	t.Json(ser.MatchReady(t.Ctx, t.Logger, t.Db, t.Nk, req))
}

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
