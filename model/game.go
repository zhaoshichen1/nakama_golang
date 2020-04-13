package model

type GameMsg struct {
	UserId string
	SessionId string
	Data *GamePlayFrame
	Point int64
}


type GamePlayFrame struct {

}