package proto

const (
	SubscribeRoom    = 5001          // 订阅房间
	SubscribeRoomAck = SubscribeRoom // 订阅房间 应答

	UnsubscribeRoom    = 5002            // 取消订阅房间
	UnsubscribeRoomAck = UnsubscribeRoom // 取消订阅房间 应答

	PushOneMessage  = 5003 // 推送单人消息
	PushRoomMessage = 5004 // 推送房间消息

	MessageAck = 5005 // 消息回应
)
