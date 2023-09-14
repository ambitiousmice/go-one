package game

type IRoom interface {
	GetRoomType() string
	OnCreated()
	OnDestroyed()
	OnJoined(player *Player)
	OnLeft(player *Player)
}
