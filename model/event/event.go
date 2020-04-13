package event

type Event string

const (
	EventMatchJoin  = Event("match_join")
	EventMatchReady = Event("match_Ready")
	EventGameRun    = Event("game_run")
)

func (v Event) String() string {
	return string(v)
}
