package proto

type JoinRoomReq struct {
	RoomID   int64
	RoomType string
}

type JoinRoomResp struct {
	RoomID int64
}
