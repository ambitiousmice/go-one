package game

import (
	"math/rand"
	"reflect"
	"sync"
)

type SceneManager struct {
	mutex             sync.RWMutex
	sceneType         string
	sceneMaxPlayerNum int
	sceneIDStart      int64
	sceneIDEnd        int64
	matchStrategy     string
	IDPool            *SceneIDPool
	scenes            map[int64]*Scene
	sceneJoinOrder    []int64
}

func NewSceneManager(sceneType string, sceneMaxPlayerNum int, sceneIDStart int64, sceneIDEnd int64, matchStrategy string) *SceneManager {
	idPool, err := NewSceneIDPool(sceneIDStart, sceneIDEnd)
	if err != nil {
		panic("init room id pool error: " + err.Error())
	}

	return &SceneManager{
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

func (sm *SceneManager) GetScene(sceneID int64) *Scene {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	return sm.scenes[sceneID]
}

func (sm *SceneManager) GetSceneByStrategy() *Scene {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	switch sm.matchStrategy {
	case SceneStrategyOrder:
		return sm.matchSceneByOrder()
	case SceneStrategyRandom:
		return sm.matchSceneRandomly()
	case SceneStrategyBalanced:
		return sm.matchSceneBalanced()
	default:
		return sm.matchSceneByOrder()
	}
}

func (sm *SceneManager) matchSceneByOrder() *Scene {

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

func (sm *SceneManager) matchSceneRandomly() *Scene {
	// 随机打乱场景加入顺序
	rand.Shuffle(len(sm.sceneJoinOrder), func(i, j int) {
		sm.sceneJoinOrder[i], sm.sceneJoinOrder[j] = sm.sceneJoinOrder[j], sm.sceneJoinOrder[i]
	})

	return sm.matchSceneByOrder()
}

func (sm *SceneManager) matchSceneBalanced() *Scene {
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

func (sm *SceneManager) createScene() *Scene {
	sceneID, err := sm.IDPool.Get()
	if err != nil {
		return nil // ID 池已用尽
	}

	sceneObjType := gameServer.getSceneObjType(sm.sceneType)
	iSceneValue := reflect.New(sceneObjType)
	iScene := iSceneValue.Interface().(IScene)

	scene := reflect.Indirect(iSceneValue).FieldByName("Scene").Addr().Interface().(*Scene)
	scene.I = iScene
	scene.init(sceneID, sm.sceneType, sm.sceneMaxPlayerNum)

	sm.scenes[scene.ID] = scene
	sm.sceneJoinOrder = append(sm.sceneJoinOrder, sceneID)

	return scene
}

func (sm *SceneManager) RemoveScene(sceneID int64) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	delete(sm.scenes, sceneID)
}

func (sm *SceneManager) GetScenes() map[int64]*Scene {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	return sm.scenes
}

func (sm *SceneManager) GetSceneCount() int {
	return len(sm.scenes)
}
