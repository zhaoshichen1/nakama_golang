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