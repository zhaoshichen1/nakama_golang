package handle

import (
	"context"
	"database/sql"
	"math/rand"
	"time"

	"nakama-golang/fantasy"
	"nakama-golang/service"
	"nakama-golang/service/game"
	"nakama-golang/service/match"

	"github.com/heroiclabs/nakama-common/runtime"
)

var (
	ser   *service.Service
	// todo match group
	mat   *match.Service
	// todo game group
	gam   *game.Service
	world = fantasy.New()
)

func rpc() {
	// todo

	world.RegistGlove("hello", helloHandle)
	world.RegistGlove("match", matchHandle)
	world.RegistGlove("match/ready", matchReady)
	world.RegistGlove("game/tick",gameTick)

	world.RegistBlade(worldEvent)
}

func initService(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule) {
	rand.Seed(time.Now().Unix())
	ser = service.New()
	mat = match.New(ctx, logger, db, nk, "hello")
	gam = game.New(ctx, logger, db, nk)
	go proxy()
}

func proxy() {
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case m := <-mat.Match:
			go gam.Start(m)
		case <-ticker.C:

		}
	}
}

func Init(ctx context.Context, logger runtime.Logger, initializer runtime.Initializer, db *sql.DB, nk runtime.NakamaModule) error {
	initService(ctx, logger, db, nk)
	rpc()
	return world.Init(initializer)
}
