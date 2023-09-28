package player

type IPlayer interface {
	OnCreated()
	OnDestroy()
	OnClientConnected()
	OnClientDisconnected()
}
