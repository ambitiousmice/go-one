package consts

import "time"

const (
	BufferedReadBufferSize  = 16384
	BufferedWriteBufferSize = 16384
)

// gate consts
const (
	GateServicePacketQueueSize = 10000
	ClientIDLength             = 19
	EntityIDLength             = 8
	ClientProxyWriteBufferSize = 1024 * 1024
	ClientProxyReadBufferSize  = 1024 * 1024

	DispatcherChannelPacketQueueSize = 4096
	GameDispatcherWriteBufferSize    = 1024 * 1024
	GameDispatcherReadBufferSize     = 1024 * 1024

	DispatcherChannelMaxTryReconnectedCount = 20
	ChannelTickInterval                     = 5 * time.Second
)

// DispatcherStatus
const (
	DispatcherStatusInit     = 1
	DispatcherStatusHealth   = 5
	DispatcherStatusUnHealth = -1
)

// DispatcherChannelStatus
const (
	DispatcherChannelStatusUnHealth      = -1
	DispatcherChannelStatusInit          = 1
	DispatcherChannelStatusHealth        = 5
	DispatcherChannelStatusStop          = 10
	DispatcherChannelStatusRestart       = 15
	DispatcherChannelStatusRestartFailed = 20
)

// game consts
const (
	GameClientPacketQueueSize  = 10000
	GameServicePacketQueueSize = 10000
	GateProxyWriteBufferSize   = 1024 * 1024
	GateProxyReadBufferSize    = 1024 * 1024
)

// IDGenerator type
const (
	Snowflake = "snowflake"
)

// server status
const (
	ServiceOffline     = 0
	ServiceTerminating = 1
	ServiceTerminated  = 5
	ServiceOnline      = 10
)

const (
	KCP_NO_DELAY                       = 1  // Whether nodelay mode is enabled, 0 is not enabled; 1 enabled
	KCP_INTERNAL_UPDATE_TIMER_INTERVAL = 10 // Protocol internal work interval, in milliseconds, such as 10 ms or 20 ms.
	KCP_ENABLE_FAST_RESEND             = 2  // Fast retransmission mode, 0 represents off by default, 2 can be set (2 ACK spans will result in direct retransmission)
	KCP_DISABLE_CONGESTION_CONTROL     = 1  // Whether to turn off flow control, 0 represents “Do not turn off” by default, 1 represents “Turn off”.

	KCP_SET_STREAM_MODE  = true
	KCP_SET_WRITE_DELAY  = true
	KCP_SET_ACK_NO_DELAY = true
)

// cron job name
const (
	CheckEnterGame = "checkEnterGame" // 检查进入游戏
)

// loginType
const (
	TokenLogin = "token"
)

// error code

const (
	ErrorCommon = 1
)

const (
	SystemError = "服务器繁忙, 请稍后再试"
	ParamError  = "param error"
)

// metadata key
const (
	Partition = "partition"
	ClusterId = "clusterID"
	WSAddr    = "wsAddr"
	TCPAddr   = "tcpAddr"
	Version   = "version"
	Status    = "status"
)

// gate kafka topic
const (
	GateBroadcastTopic = "gate-broadcast"
)

const (
	GateBroadcastTypeNotice = "notice"
)
