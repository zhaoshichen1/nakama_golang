package handle

import (
	"encoding/json"
	"strconv"

	"github.com/heroiclabs/nakama-common/runtime"

	"nakama-golang/fantasy"
	"nakama-golang/model"
	"nakama-golang/model/event"
)

func ToInt64(str string) int64 {
	v, _ := strconv.ParseInt(str, 10, 64)
	return v
}

func worldEvent(c *fantasy.Claude) {
	switch c.Event() {
	case event.EventMatchJoin:
		userId, ok := c.Ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if !ok {
			// User ID not found in the context.
			return
		}
		info := c.Evt.Properties
		aid := ToInt64(info["aid"])
		matchManager.NewPlayer(aid, []string{userId})
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
		aid := ToInt64(info["aid"])
		matchId := info["match_id"]
		matchManager.ReadyMatch(aid, matchId, userId, sessionID)
	case event.EventGameReady:
		info := c.Evt.Properties
		msg := &model.GameMsg{
			UserId:    info["user_id"],
			SessionId: info["session_id"],
			MatchId:   info["match_id"],
			Data:      nil,
		}
		gameGroup.Run(msg)
	case event.EventGameRun:
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
		data := &model.GamePlayFrame{}
		json.Unmarshal([]byte(info["data"]), data)
		matchId, ok := c.Ctx.Value(runtime.RUNTIME_CTX_MATCH_ID).(string)
		if !ok {
			return
		}
		msg := &model.GameMsg{
			UserId:    userId,
			SessionId: sessionID,
			MatchId:   matchId,
			Data:      data,
		}
		gameGroup.Run(msg)
	}
}
