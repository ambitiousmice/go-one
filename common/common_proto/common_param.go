package common_proto

type ErrorResp struct {
	Code int32
	Data string
}

type EnterGameFromServerParam struct {
	ClientID string
}

type EnterGameReq struct {
	AccountType string
	Account     string
	EntityID    int64
	ClientID    string
	Game        string
}

type EnterGameResp struct {
	EntityID int64
	Game     string
}

type GameDispatcherChannelInfoReq struct {
	GateClusterID uint8
	Game          string
	GameClusterID uint8
	ChannelID     uint8
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
