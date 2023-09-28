package scene_center

import (
	"go-one/common/log"
	"go-one/game/common"
	"go-one/game/player"
	"math/rand"
	"reflect"
	"sync"
)

var ManagerContext = make(map[string]*Manager)
var sceneTypes = make(map[string]reflect.Type)

type Manager struct {
	mutex             sync.RWMutex
	sceneType         string
	sceneMaxPlayerNum int
	sceneIDStart      int64
	sceneIDEnd        int64
	matchStrategy     string
	IDPool            *IDPool
	scenes            map[int64]*Scene
	sceneJoinOrder    []int64
}

func NewSceneManager(sceneType string, sceneMaxPlayerNum int, sceneIDStart int64, sceneIDEnd int64, matchStrategy string) *Manager {
	idPool, err := NewIDPool(sceneIDStart, sceneIDEnd)
	if err != nil {
		panic("init room id pool error: " + err.Error())
	}

	return &Manager{
		sceneType:         sceneType,
		sceneIDStart:      sceneIDStart,
		sceneIDEnd:        sceneIDEnd,
		matchStrategy:     matchStrategy,
		sceneMaxPlayerNum: sceneMaxPlayerNum,
		IDPool:            idPool,
		scenes:            make(map[int64]*Scene),
		sceneJoinOrder:    make([]int64, 0),
	}
}

func (sm *Manager) GetScene(sceneID int64) *Scene {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	return sm.scenes[sceneID]
}

func (sm *Manager) GetSceneByStrategy() *Scene {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	switch sm.matchStrategy {
	case common.SceneStrategyOrder:
		return sm.matchSceneByOrder()
	case common.SceneStrategyRandom:
		return sm.matchSceneRandomly()
	case common.SceneStrategyBalanced:
		return sm.matchSceneBalanced()
	default:
		return sm.matchSceneByOrder()
	}
}

func (sm *Manager) matchSceneByOrder() *Scene {

	// 遍历房间加入的顺序
	for _, sceneID := range sm.sceneJoinOrder {
		scene := sm.scenes[sceneID]
		if scene != nil && scene.GetPlayerCount() < scene.MaxPlayerNum {
			return scene
		}
	}

	// 如果没有可用房间，创建一个新房间
	return sm.createScene()
}

func (sm *Manager) matchSceneRandomly() *Scene {
	// 随机打乱场景加入顺序
	rand.Shuffle(len(sm.sceneJoinOrder), func(i, j int) {
		sm.sceneJoinOrder[i], sm.sceneJoinOrder[j] = sm.sceneJoinOrder[j], sm.sceneJoinOrder[i]
	})

	return sm.matchSceneByOrder()
}

func (sm *Manager) matchSceneBalanced() *Scene {
	// 寻找最少人数的场景
	var minScene *Scene
	minPlayers := 99999 // 初始值设置为一个较大的数

	for _, sceneID := range sm.sceneJoinOrder {
		room := sm.scenes[sceneID]
		if room != nil {
			playerCount := room.GetPlayerCount()
			if playerCount < room.MaxPlayerNum && playerCount < minPlayers {
				minPlayers = playerCount
				minScene = room
			}
		}
	}

	// 如果没有可用房间，创建一个新房间
	if minScene == nil {
		return sm.createScene()
	}

	return minScene
}

func (sm *Manager) createScene() *Scene {
	sceneID, err := sm.IDPool.Get()
	if err != nil {
		log.Warnf("create scene error: %s", err.Error())
		return nil // ID 池已用尽
	}

	sceneObjType := getSceneObjType(sm.sceneType)
	iSceneValue := reflect.New(sceneObjType)
	iScene := iSceneValue.Interface().(IScene)

	scene := reflect.Indirect(iSceneValue).FieldByName("Scene").Addr().Interface().(*Scene)
	scene.I = iScene
	scene.init(sceneID, sm.sceneType, sm.sceneMaxPlayerNum)

	sm.scenes[scene.ID] = scene
	sm.sceneJoinOrder = append(sm.sceneJoinOrder, sceneID)

	return scene
}

func (sm *Manager) RemoveScene(sceneID int64) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	delete(sm.scenes, sceneID)
}

func (sm *Manager) GetScenes() map[int64]*Scene {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	return sm.scenes
}

func (sm *Manager) GetSceneCount() int {
	return len(sm.scenes)
}

// RegisterSceneType register a scene_center type
func RegisterSceneType(scene IScene) {
	if sceneTypes[scene.GetSceneType()] != nil {
		panic("scene type already registered, sceneType:" + scene.GetSceneType())
	}

	objVal := reflect.ValueOf(scene)
	objType := objVal.Type()

	if objType.Kind() == reflect.Ptr {
		objType = objType.Elem()
	}

	sceneTypes[scene.GetSceneType()] = objType

	log.Infof("register scene: %s", scene.GetSceneType())
}

func getSceneObjType(sceneType string) reflect.Type {
	objType := sceneTypes[sceneType]
	if objType == nil {
		panic("scene type not found, sceneType:" + sceneType)
	}

	return objType
}

func GetSceneManager(sceneType string) *Manager {
	sceneManager := ManagerContext[sceneType]

	if sceneManager == nil {
		panic("scene manager not found, sceneType:" + sceneType)
	}

	return sceneManager
}

func GetSceneByPlayer(player *player.Player) *Scene {
	manager := GetSceneManager(player.SceneType)
	if manager == nil {
		return nil
	}

	return manager.GetScene(player.SceneID)
}

func JoinScene(sceneType string, sceneID int64, player *player.Player) {
	sceneManager := GetSceneManager(sceneType)

	var scene *Scene
	if sceneID == 0 {
		scene = sceneManager.GetSceneByStrategy()
	} else {
		scene = sceneManager.GetScene(sceneID)
	}

	if scene == nil {
		player.SendCommonErrorMsg(common.ServerIsFull)
	}

	scene.join(player)
}

func ReJoinScene(player *player.Player) {
	scene := GetSceneByPlayer(player)
	if scene == nil {
		player.SendCommonErrorMsg(common.ServerIsFull)
	}

	scene.join(player)
}

func Leave(player *player.Player) {
	scene := GetSceneByPlayer(player)
	if scene == nil {
		return
	}

	scene.leave(player)
}
