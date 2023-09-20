package proto

type SubscribeRoomReq struct {
	RoomID int64
}

type SubscribeRoomResp struct {
	RoomID int64
}

type UnSubscribeRoomReq struct {
	RoomID int64
}

type UnSubscribeRoomResp struct {
	RoomID int64
}

type PushMessageReq struct {
	RoomID int64
	Msg    string
}

type ChatMessage struct {
	RoomID int64
	From   int64
	Msg    string
}
