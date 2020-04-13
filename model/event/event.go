package event

type Event string

const (
	// match
	EventMatchJoin  = Event("match_join")
	EventMatchReady = Event("match_Ready")
	// game
	EventGameReady = Event("game_ready")
	EventGameRun   = Event("game_run")
)

func (v Event) String() string {
	return string(v)
}
