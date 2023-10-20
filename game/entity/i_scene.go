package entity

type IScene interface {
	GetSceneType() string
	OnCreated()
	OnDestroyed()
	OnJoined(player *Player)
	OnLeft(player *Player)
}
