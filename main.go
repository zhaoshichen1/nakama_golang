package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/pkg/errors"

	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/runtime"
)

func main() {}

func InitModule(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error {
	logger.Info("start InitModule")
	// Register the RPC function.
	if err := initializer.RegisterRpc("hello", Hello); err != nil {
		logger.Error("Unable to register: %v", err)
		return err
	}
	if err:=initializer.RegisterRpc("bar",Bar);err!=nil{
		return err
	}
	if err:=initializer.RegisterEvent(processEvent);err!=nil{
		return err
	}
	logger.Info("InitModule success")
	return nil
}

func processEvent(ctx context.Context, logger runtime.Logger, evt *api.Event){
	switch evt.GetName() {
	case "bar":
		logger.Info("process evt: %+v", evt)
	case "hello":
		logger.Info("process evt: %+v", evt)
	default:
		logger.Error("unrecognised evt: %+v", evt)
	}
}

func Bar(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string)(string,error){
	evt := &api.Event{
		Name:       "bar",
		Properties: map[string]string{
			"my_key": "my_value",
		},
		External:   true,
	}
	if err := nk.Event(ctx, evt); err != nil {
		// Handle error.
		logger.Error("Event err:%+v",err)
		return "",err
	}
	return "success",nil
}

func Hello(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	logger.Info("call hello : %s", []byte(payload))
	userId, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	if !ok {
		// User ID not found in the context.
		return "", errors.New("userId not found!")
	}
	val:=map[string]interface{}{}
	if err:=json.Unmarshal([]byte(payload),&val);err!=nil{
		return "",errors.Wrapf(err,"Hello json.Unmarshal")
	}
	logger.Info("val:%+v",val)
	logger.Info("userid:%v", userId)
	event:=&api.Event{Name:"hello"}
	if err:=nk.Event(ctx,event);err!=nil{
		return "",err
	}
	if err := nk.NotificationSend(ctx, userId, "233", map[string]interface{}{"hello": payload}, 1, "", false); err != nil {
		logger.Error("NotificationSend err:%v", err)
		return "", errors.Wrapf(err,"Hello NotificationSend userid:%v",userId)
	}
	return "Success", nil
}
