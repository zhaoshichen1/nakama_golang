package handle

import (
	"github.com/heroiclabs/nakama-common/runtime"
	"nakama-golang/fantasy"
	"nakama-golang/model/event"
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

func worldEvent(c *fantasy.Claude) {
	switch c.Event() {
	case event.EventMatchJoin:
		userId, ok := c.Ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if !ok {
			// User ID not found in the context.
			return
		}
		mat.AddPlayer(userId)
	case event.EventMatchReady:
		info := c.Evt.Properties
		userId, ok := c.Ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if !ok {
			// User ID not found in the context.
			return
		}
		sessionID, ok := c.Ctx.Value(runtime.RUNTIME_CTX_SESSION_ID).(string)
		if !ok {
			// If session ID is not found, RPC was not called over a connected socket.
			return
		}
		mat.ReadyMatch(info["match_id"], userId, sessionID)
	}
}
