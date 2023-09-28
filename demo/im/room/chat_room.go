package room

import (
	"fmt"
	"github.com/robfig/cron/v3"
	"go-one/common/log"
	"go-one/demo/im/proto"
	"go-one/game/player"
	"sync"
	"sync/atomic"
	"time"
)

var messageCount uint64
var cronTab = cron.New(cron.WithSeconds())

func init() {
	cronTab.Start()
	cronTab.AddFunc("@every 1s", func() {
		log.Infof("message count: %d", atomic.LoadUint64(&messageCount))
		atomic.StoreUint64(&messageCount, 0)
	})
}

type ChatRoom struct {
	ID   int64
	name string

	pMutex  sync.RWMutex
	players map[int64]*player.Player

	rwMutex         sync.RWMutex
	broadcastTimer  *time.Timer
	msgBuffer       []*proto.ChatMessage
	msgBufferMaxLen int
}

func NewChatRoom(id int64, name string) *ChatRoom {
	room := &ChatRoom{
		ID:              id,
		name:            name,
		players:         make(map[int64]*player.Player),
		broadcastTimer:  time.NewTimer(time.Millisecond * 50),
		msgBufferMaxLen: 1024,
	}

	go room.broadcastTask()

	return room
}

func (r *ChatRoom) String() string {
	return fmt.Sprintf("room info:ID=<%d>", r.ID)

}
func (r *ChatRoom) Join(player *player.Player) {
	r.pMutex.Lock()
	defer r.pMutex.Unlock()

	r.players[player.EntityID] = player

	log.Infof("%s join %s", player, r)
}

func (r *ChatRoom) Leave(p *player.Player) {
	r.pMutex.Lock()
	defer r.pMutex.Unlock()

	delete(r.players, p.EntityID)
}

func (r *ChatRoom) Broadcast(msg *proto.ChatMessage) {
	r.rwMutex.Lock()
	defer r.rwMutex.Unlock()
	// 将消息添加到缓冲区
	r.msgBuffer = append(r.msgBuffer, msg)
	if len(r.msgBuffer) >= r.msgBufferMaxLen {
		r.sendMessages()
	}

}

func (r *ChatRoom) broadcastTask() {
	for {
		select {
		case <-r.broadcastTimer.C:
			r.rwMutex.Lock()
			// 检查缓冲区是否有消息
			if len(r.msgBuffer) > 0 {
				r.sendMessages()
			}
			r.broadcastTimer.Reset(time.Millisecond * 50)
			r.rwMutex.Unlock()
		}
	}
}

func (r *ChatRoom) sendMessages() {
	// 复制消息并清空缓冲区
	messages := make([]*proto.ChatMessage, len(r.msgBuffer))
	copy(messages, r.msgBuffer)
	r.msgBuffer = r.msgBuffer[:0]

	// 发送消息给所有玩家
	r.pMutex.RLock()
	for _, p := range r.players {
		p.SendGameData(proto.MessageAck, messages)
	}
	r.pMutex.RUnlock()

	// 原子增加消息计数
	atomic.AddUint64(&messageCount, uint64(len(messages)))
}
