package proto

type ErrorResp struct {
	Code int32
	Msg  string
}

type LoginReq struct {
	LoginType string
	Account   string
	Game      string
}

type LoginResp struct {
	EntityID int64 `json:"entityID"`
}

type GameDispatcherChannelInfoReq struct {
	GateID    uint8
	Game      string
	GameID    uint8
	ChannelID uint8
}

type GameDispatcherChannelInfoResp struct {
	Success bool
	Msg     string
}

type NewPlayerConnectionReq struct {
	ClientID string
	EntityID int64
}

type PlayerDisconnectedReq struct {
	ClientID string
	EntityID int64
}

type GameReq struct {
	Cmd   uint16
	Param []byte
}

type GameResp struct {
	Cmd  uint16
	Code int32
	Data []byte
}
