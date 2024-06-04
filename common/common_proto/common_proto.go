package common_proto

// 服务端协议 uint16

const (
	Error = 1 // 全局异常
)

// from server
const (
	ConnectionSuccessFromServer = 1001
	OfflineFromServer           = 1002
	BroadcastFromServer         = 1003
)

// from client
const (
	HeartbeatFromClient    = 2001                // 心跳
	HeartbeatFromClientAck = HeartbeatFromClient // 心跳应答

	OfflineFromClient = 2002 //客户端主动下线

	LoginFromClient    = 2003            // 客户端登录
	LoginFromClientAck = LoginFromClient // 客户端登录应答

)

// game dispatcher
const (
	HeartbeatFromDispatcher    = 3001 // 游戏调度器心跳
	HeartbeatFromDispatcherAck = HeartbeatFromDispatcher

	GameDispatcherChannelInfoFromDispatcher    = 3002                                    // 发送游戏调度器通道信息
	GameDispatcherChannelInfoFromDispatcherAck = GameDispatcherChannelInfoFromDispatcher // 发送游戏调度器通道信息 应答

	NewPlayerConnectionFromDispatcher = 3003 // 新玩家连接

	PlayerDisconnectedFromDispatcher = 3004 // 玩家断开连接
)

// game

const (
	JoinScene    = 4001      // 加入场景
	JoinSceneAck = JoinScene // 加入场景 应答

	LeaveScene    = 4002       // 离开场景
	LeaveSceneAck = LeaveScene // 离开场景 应答

	CreateEntity = 4003 //创建玩家

	DestroyEntity = 4004 //销毁玩家

	Move = 4005

	AOISync = 4006
)
