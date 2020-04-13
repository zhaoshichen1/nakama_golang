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
	ser        *service.Service
	matchGroup *match.Group
	gameGroup  *game.Group
	world      = fantasy.New()
)

func rpc() {
	// todo

	world.RegistGlove("hello", helloHandle) // 测试接口
	world.RegistGlove("match", matchHandle)
	world.RegistGlove("match/ready", matchReady)
	world.RegistGlove("game/tick", gameTick)

	world.RegistBlade(worldEvent)
}

func initService(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule) {
	rand.Seed(time.Now().Unix())
	ser = service.New()
	matchGroup = match.NewGroup()
	matchGroup.Add(match.New(ctx, logger, db, nk, 223344, "game"))
	gameGroup = game.NewGroup(ctx, logger, db, nk)
	go gameGroup.Tick()
	go proxy()
}

func proxy() {
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case m := <-matchGroup.Match:
			go gameGroup.Start(m)
		case <-ticker.C:

		}
	}
}

func Init(ctx context.Context, logger runtime.Logger, initializer runtime.Initializer, db *sql.DB, nk runtime.NakamaModule) error {
	initService(ctx, logger, db, nk)
	rpc()
	return world.Init(initializer)
}
