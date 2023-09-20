package game

import (
	"github.com/robfig/cron/v3"
	"sync"
)

type BaseScene struct {
	mutex        sync.RWMutex
	ID           int64
	Name         string
	Type         string
	MaxPlayerNum int
	players      map[int64]*Player
	cron         *cron.Cron
	cronTaskMap  map[string]cron.EntryID
}

func NewBaseScene(id int64, sceneType string, maxPlayerNum int) *BaseScene {
	return &BaseScene{
		ID:           id,
		Type:         sceneType,
		MaxPlayerNum: maxPlayerNum,
		players:      map[int64]*Player{},
		cron:         cron.New(cron.WithSeconds()),
		cronTaskMap:  map[string]cron.EntryID{},
	}
}

func (br *BaseScene) GetPlayer(entityID int64) *Player {
	br.mutex.RLock()
	defer br.mutex.RUnlock()

	return br.players[entityID]
}

func (br *BaseScene) AddPlayer(player *Player) {
	br.mutex.Lock()
	defer br.mutex.Unlock()

	br.players[player.EntityID] = player
}

func (br *BaseScene) RemovePlayer(player *Player) {
	br.mutex.Lock()
	defer br.mutex.Unlock()

	delete(br.players, player.EntityID)
}

func (br *BaseScene) GetPlayerCount() int {
	return len(br.players)
}

func (br *BaseScene) AddCronTask(taskName string, spec string, method func()) error {
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

func (br *BaseScene) RemoveCronTask(taskName string) {
	br.mutex.Lock()
	defer br.mutex.Unlock()

	taskID := br.cronTaskMap[taskName]
	if taskID != 0 {
		br.cron.Remove(taskID)
	}
}

func (br *BaseScene) PushOne(entityID int64, cmd uint16, data interface{}) {
	br.mutex.RLock()
	defer br.mutex.RUnlock()

	player := br.players[entityID]
	if player == nil {
		return
	}

	player.SendGameData(cmd, data)
}

func (br *BaseScene) Broadcast(cmd uint16, data interface{}) {
	br.mutex.RLock()
	defer br.mutex.RUnlock()

	for _, player := range br.players {
		player.SendGameData(cmd, data)
	}

}
