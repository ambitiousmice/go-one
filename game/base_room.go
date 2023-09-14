package game

import (
	"github.com/robfig/cron/v3"
	"go-one/common/proto"
	"sync"
)

type BaseRoom struct {
	mutex        sync.RWMutex
	ID           int64
	Name         string
	Type         string
	MaxPlayerNum int
	players      map[int64]*Player
	cron         *cron.Cron
	cronTaskMap  map[string]cron.EntryID
}

func NewBaseRoom(id int64, roomType string, maxPlayerNum int) *BaseRoom {
	return &BaseRoom{
		ID:           id,
		Type:         roomType,
		MaxPlayerNum: maxPlayerNum,
		players:      map[int64]*Player{},
		cron:         cron.New(cron.WithSeconds()),
		cronTaskMap:  map[string]cron.EntryID{},
	}
}

func (br *BaseRoom) GetPlayer(entityID int64) *Player {
	br.mutex.RLock()
	defer br.mutex.RUnlock()

	return br.players[entityID]
}

func (br *BaseRoom) AddPlayer(player *Player) {
	br.mutex.Lock()
	defer br.mutex.Unlock()

	br.players[player.entityID] = player
}

func (br *BaseRoom) RemovePlayer(player *Player) {
	br.mutex.Lock()
	defer br.mutex.Unlock()

	delete(br.players, player.entityID)
}

func (br *BaseRoom) GetPlayerCount() int {
	return len(br.players)
}

func (br *BaseRoom) AddCronTask(taskName string, spec string, method func()) error {
	br.mutex.Lock()
	defer br.mutex.Unlock()

	taskID := br.cronTaskMap[taskName]
	if taskID != 0 {
		br.cron.Remove(taskID)
	}

	newTaskID, err := br.cron.AddFunc(spec, method)
	if err != nil {
		return err
	}

	br.cronTaskMap[taskName] = newTaskID

	return nil
}

func (br *BaseRoom) RemoveCronTask(taskName string) {
	br.mutex.Lock()
	defer br.mutex.Unlock()

	taskID := br.cronTaskMap[taskName]
	if taskID != 0 {
		br.cron.Remove(taskID)
	}
}

func (br *BaseRoom) PushOne(entityID int64, msg *proto.GameResp) {
	br.mutex.RLock()
	defer br.mutex.RUnlock()

	player := br.players[entityID]
	if player == nil {
		return
	}

	player.SendGameMsg(msg)
}

func (br *BaseRoom) Broadcast(msg *proto.GameResp) {
	br.mutex.RLock()
	defer br.mutex.RUnlock()

	for _, player := range br.players {
		player.SendGameMsg(msg)
	}

}
