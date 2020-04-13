package handle

import (
	"encoding/json"

	"nakama-golang/fantasy"
	"nakama-golang/model"
	"nakama-golang/model/event"

	"github.com/heroiclabs/nakama-common/runtime"
)

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
	case event.EventGameRun:
		info:=c.Evt.Properties
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
		data:=&model.GamePlayFrame{}
		json.Unmarshal([]byte(info["data"]),data)
		msg:=&model.GameMsg{
			UserId:userId,
			SessionId:sessionID,
			Data:data,
		}
		gam.Run(msg)
	}
}

