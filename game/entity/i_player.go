package entity

type IPlayer interface {
	OnCreated()
	OnDestroy()
	OnClientDisconnected()
	UpdateData()
}
