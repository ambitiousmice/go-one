package common_proto

//
//import (
//	"go-one/game/common"
//)
//
//type ErrorResp struct {
//	Code int32
//	Data string
//}
//
//type ConnectionSuccessFromServerResp struct {
//	ClientID string
//}
//
//type LoginReq struct {
//	AccountType string
//	Account     string
//	EntityID    int64
//	ClientID    string
//	Game        string
//	Region      int32
//}
//
//type LoginResp struct {
//	EntityID int64
//	Game     string
//	Region   int32
//}
//
//type GameDispatcherChannelInfoReq struct {
//	GateClusterID uint8
//	Game          string
//	GameClusterID uint8
//	ChannelID     uint8
//}
//
//type GameDispatcherChannelInfoResp struct {
//	Success bool
//	Msg     string
//}
//
//type NewPlayerConnectionReq struct {
//	EntityID int64
//	Region   int32
//}
//
//type PlayerDisconnectedReq struct {
//	EntityID int64
//}
//
//type GameReq struct {
//	Cmd   uint16
//	Param []byte
//}
//
//type GameResp struct {
//	Cmd  uint16
//	Code int32
//	Data []byte
//}
//
//type JoinSceneReq struct {
//	SceneType string
//	SceneID   int64
//}
//
//type JoinSceneResp struct {
//	SceneType string
//	SceneID   int64
//}
//
//type LeaveSceneReq struct {
//	SceneType string
//	SceneID   int64
//}
//
//type LeaveSceneResp struct {
//	SceneType string
//	SceneID   int64
//}
//
//type OnCreateEntity struct {
//	EntityID int64
//	Type     string
//	X        common.Coord   /*`json:"x"`*/
//	Y        common.Coord   /*`json:"y"`*/
//	Z        common.Coord   /*`json:"z"`*/
//	Yaw      common.Yaw     /*`json:"Yaw"`*/
//	Speed    common.Speed   /*`json:"Speed"`*/
//	Attr     map[string]any /*`json:"attr"`*/
//}
//
//type OnDestroyEntity struct {
//	EntityID int64
//}
//
//type AOISyncInfo struct {
//	EntityID int64
//	X        common.Coord /*`json:"x"`*/
//	Y        common.Coord /*`json:"y"`*/
//	Z        common.Coord /*`json:"z"`*/
//	Yaw      common.Yaw   /*`json:"Yaw"`*/
//	Speed    common.Speed /*`json:"Speed"`*/
//}
//
//type MoveReq struct {
//	X     common.Coord
//	Y     common.Coord
//	Z     common.Coord
//	Yaw   common.Yaw
//	Speed common.Speed
//}
//
//type GateBroadcastMsg struct {
//	Type string
//	Data []byte
//}
