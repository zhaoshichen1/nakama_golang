package event

type Event string

const (
	EventMatchJoin  = Event("match_join")
	EventMatchReady = Event("match_Ready")
)

func (v Event) String() string {
	return string(v)
}
