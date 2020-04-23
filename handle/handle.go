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
	ser          *service.Service
	matchManager *match.Manager
	gameGroup    *game.Group
	world        = fantasy.New()
)

func rpc() {
	// todo

	world.RegistGlove("hello", helloHandle) // 测试接口
	world.RegistGlove("match", matchHandle)
	world.RegistGlove("match/ready", matchReady)
	world.RegistGlove("game/ready", gameReady)
	world.RegistGlove("game/tick", gameTick)

	world.RegistBlade(worldEvent)
}

func initService(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule) {
	rand.Seed(time.Now().Unix())
	ser = service.New()

	matchManager = match.NewMatchManager(ctx)
	go matchManager.Match() // 异步定时轮询匹配

	gameGroup = game.NewGroup(ctx, logger, db, nk)
	go gameGroup.Tick()
	go proxy()
}

func proxy() {
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case m := <-matchManager.NewMatch:
			go gameGroup.Start(m)
		case <-ticker.C:

		}
	}
}

func MakeMatch(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, entries []runtime.MatchmakerEntry) (string, error) {

	for _, e := range entries {
		logger.Info("Matched user '%s' named '%s'", e.GetPresence().GetUserId(), e.GetPresence().GetUsername())
		for k, v := range e.GetProperties() {
			logger.Info("Matched on '%s' value '%v'", k, v)
		}
	}

	matchId, err := nk.MatchCreate(ctx, "dance_battle", map[string]interface{}{"invited": entries})
	if err != nil {
		logger.Error("CreateNewMatch Err %v", err)
		return "", err
	}
	logger.Info("New MatchID ", matchId)

	return matchId, nil
}

func Init(ctx context.Context, logger runtime.Logger, initializer runtime.Initializer, db *sql.DB, nk runtime.NakamaModule) error {
	initService(ctx, logger, db, nk)
	rpc()

	// 注册匹配命中逻辑
	if err := initializer.RegisterMatchmakerMatched(MakeMatch); err != nil {
		logger.Error("Unable to register: %v", err)
		return err
	}

	return world.Init(initializer)
}
