package proto

type SubscribeRoomReq struct {
	RoomID int64
}

type SubscribeRoomResp struct {
	RoomID int64
}

type UnsubscribeRoomReq struct {
	RoomID int64
}

type UnsubscribeRoomResp struct {
	RoomID int64
}

type PushMessageReq struct {
	RoomID int64
	From   int64
	To     int64
	Msg    string
}

type ChatMessage struct {
	RoomID int64
	From   int64
	Msg    string
}
