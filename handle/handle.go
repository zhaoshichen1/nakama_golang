package handle

import (
	"nakama-golang/fantasy"
	"nakama-golang/service"

	"github.com/heroiclabs/nakama-common/runtime"
)

var (
	ser = service.New()
	world = fantasy.New()
)

func rpc() {
	// todo
	world.RegistGlove("hello", helloHandle)

	world.RegistBlade(helloEvent,worldEvent)
}

func Init(initializer runtime.Initializer) error {

	rpc()
	return world.Init(initializer)
}
