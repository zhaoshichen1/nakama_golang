package main

import (
	"context"
	"database/sql"

	"nakama-golang/handle"

	"github.com/heroiclabs/nakama-common/runtime"
)

func main() {}

func InitModule(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error {
	logger.Info("start InitModule")
	if err := handle.Init(ctx, logger, initializer, db, nk); err != nil {
		logger.Error("InitModule err:%+v", err)
		return err
	}
	logger.Info("InitModule success")
	return nil
}
