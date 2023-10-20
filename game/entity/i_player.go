package entity

type IPlayer interface {
	OnCreated()
	OnDestroy()
	OnClientConnected()
	OnClientDisconnected()
	OnJoinScene()
}
