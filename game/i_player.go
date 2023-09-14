package game

type IPlayer interface {
	OnCreated()
	OnDestroy()
	OnClientConnected()
	OnClientDisconnected()
}
