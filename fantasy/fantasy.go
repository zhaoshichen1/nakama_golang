package fantasy

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/runtime"
	"nakama-golang/model/event"
)

type Glove func(tifa *Tifa)
type Gloves []Glove

type Blade func(claude *Claude)
type Blades []Blade

type World struct {
	Heroine map[string]Gloves // rpc方法
	Hero    Blades            // event响应方法
}

func New() *World {
	return &World{
		Heroine: map[string]Gloves{},
	}
}

func (v *World) RegistGlove(hash string, handle ...Glove) {
	v.Heroine[hash] = append(v.Heroine[hash], handle...)
}

func (v *World) RegistBlade(handle ...Blade) {
	v.Hero = append(v.Hero, handle...)
}

func (v *World) Init(initializer runtime.Initializer) error {
	for k, v := range v.Heroine {
		if err := initializer.RegisterRpc(k, heroine(v)); err != nil {
			return err
		}
	}
	if err := initializer.RegisterEvent(hero(v.Hero)); err != nil {
		return err
	}
	return nil
}

type Claude struct {
	Ctx    context.Context
	Logger runtime.Logger
	Evt    *api.Event
	// my
	Blades  Blades
	isAbort bool
}

func (c *Claude) Event() event.Event {
	return event.Event(c.Evt.Name)
}

func (c *Claude) Abort() {
	c.isAbort = true
}

type Tifa struct {
	Ctx    context.Context
	Logger runtime.Logger
	Db     *sql.DB
	Nk     runtime.NakamaModule
	// my
	Gloves   Gloves
	isAbort  bool
	response string
	request  string
	err      error
}

func (c *Tifa) Bind(v interface{}) error {
	if err := json.Unmarshal([]byte(c.request), v); err != nil {
		c.err = err
		return err
	}
	return nil
}

func (c *Tifa) Abort() {
	c.isAbort = true
}

func (c *Tifa) Json(v interface{}, err error) {
	jstr, err := json.Marshal(v)
	c.response = string(jstr)
	c.err = err
}

func (c *Tifa) String(str string, err error) {
	c.response = str
	c.err = err
}

func heroine(handles Gloves) func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	tifa := &Tifa{}
	tifa.Gloves = handles
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (s string, err error) {
		tifa.Ctx = ctx
		tifa.Logger = logger
		tifa.Db = db
		tifa.Nk = nk
		tifa.request = payload
		for i := range tifa.Gloves {
			tifa.Gloves[i](tifa)
			if tifa.isAbort {
				return tifa.response, tifa.err
			}
		}
		return tifa.response, tifa.err
	}
}

func hero(blades Blades) func(ctx context.Context, logger runtime.Logger, evt *api.Event) {
	claude := &Claude{}
	claude.Blades = blades
	return func(ctx context.Context, logger runtime.Logger, evt *api.Event) {
		claude.Ctx = ctx
		claude.Logger = logger
		claude.Evt = evt
		for i := range claude.Blades {
			claude.Blades[i](claude)
			if claude.isAbort {
				return
			}
		}
	}
}
