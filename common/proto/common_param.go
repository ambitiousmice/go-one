package proto

type ErrorResp struct {
	Code int32
	Msg  string
}

type EnterGameFromServerParam struct {
	ClientID string
}

type EnterGameReq struct {
	AccountType  string
	Account      string
	Reconnection bool
	EntityID     int64
	ClientID     string
	Game         string
	GameID       uint8
}

type EnterGameResp struct {
	EntityID int64 `json:"entityID"`
	Game     string
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
	EntityID int64
}

type PlayerDisconnectedReq struct {
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

type JoinSceneReq struct {
	SceneType string
	SceneID   int64
}

type JoinSceneResp struct {
	SceneType string
	SceneID   int64
}

type LeaveSceneReq struct {
	SceneType string
	SceneID   int64
}

type LeaveSceneResp struct {
	SceneType string
	SceneID   int64
}
