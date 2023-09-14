package game

type Room struct {
	*BaseRoom

	I IRoom
}

func (r *Room) init(id int64, roomType string, maxPlayerNum int) {
	r.BaseRoom = NewBaseRoom(id, roomType, maxPlayerNum)

	r.I.OnCreated()
}

func (r *Room) Join(player *Player) {
	r.AddPlayer(player)

	r.I.OnJoined(player)
}
