package handle

import (
	"nakama-golang/fantasy"
	"nakama-golang/model"
)

func helloHandle(t *fantasy.Tifa) {
	req := &model.ReqHello{}
	if err := t.Bind(req); err != nil {
		t.Abort()
		return
	}
	res, err := ser.Hello(t.Ctx, t.Logger, t.Db, t.Nk, req)
	t.Json(res, err)
}

func helloEvent(c *fantasy.Claude) {

}

func worldEvent(c *fantasy.Claude) {

}
