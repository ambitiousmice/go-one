package entity

import (
	"github.com/ambitiousmice/go-one/common/context"
	"github.com/ambitiousmice/go-one/common/log"
	"github.com/ambitiousmice/go-one/common/pool/goroutine_pool"
	"github.com/ambitiousmice/go-one/game/common"
	"math/rand"
	"reflect"
	"sync"
	"time"
)

var SceneManagerContext = make(map[string]*SceneManager)
var sceneTypes = make(map[string]reflect.Type)
var playerCountMap = make(map[string]int)

var sceneMsgChan = make(chan func(), 102400)

func init() {
	err := context.AddCronTask("scene_player_count_task", "0 0/1 * * * ?", func() {
		for _, manager := range SceneManagerContext {
			playerCount := 0
			for _, scene := range manager.scenes {
				playerCount += scene.GetPlayerCount()
			}
			playerCountMap[manager.sceneType] = playerCount
		}
		// TODO: 上报监控

	})

	if err != nil {
		log.Panic("add cronTab task scene_player_count_task error: " + err.Error())
	}

	go func() {
		for {
			select {
			case task := <-sceneMsgChan:
				errs := goroutine_pool.Submit(task)
				if errs != nil {
					log.Warnf("submit scene msg err:%s", errs.Error())
				}

				//log.Infof("process scene aoiMsg task")
			}
		}
	}()
}

type SceneManager struct {
	mutex             sync.RWMutex
	sceneType         string
	sceneMaxPlayerNum int
	sceneIDStart      int64
	sceneIDEnd        int64
	matchStrategy     string
	IDPool            *IDPool
	scenes            map[int64]*Scene
	sceneJoinOrder    []int64
	enableAOI         bool
	aoiDistance       float32
	tickRate          time.Duration
}

func NewSceneManager(sceneType string, sceneMaxPlayerNum int, sceneIDStart int64, sceneIDEnd int64, matchStrategy string, enableAOI bool, aoiDistance float32, tickRate time.Duration) *SceneManager {
	idPool, err := NewIDPool(sceneIDStart, sceneIDEnd)
	if err != nil {
		log.Panic("init room id pool error: " + err.Error())
	}

	if tickRate == 0 {
		tickRate = 34 * time.Millisecond
	}

	m := &SceneManager{
		sceneType:         sceneType,
		sceneIDStart:      sceneIDStart,
		sceneIDEnd:        sceneIDEnd,
		matchStrategy:     matchStrategy,
		sceneMaxPlayerNum: sceneMaxPlayerNum,
		IDPool:            idPool,
		scenes:            make(map[int64]*Scene),
		sceneJoinOrder:    make([]int64, 0),
		enableAOI:         enableAOI,
		aoiDistance:       aoiDistance,
		tickRate:          tickRate,
	}

	if enableAOI {
		m.syncAOIInfoTicker()
	}

	return m
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
	case common.SceneStrategyOrder:
		return sm.matchSceneByOrder()
	case common.SceneStrategyRandom:
		return sm.matchSceneRandomly()
	case common.SceneStrategyBalanced:
		return sm.matchSceneBalanced()
	case common.SceneStrategyMax:
		return sm.matchSceneMax()
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

func (sm *SceneManager) matchSceneMax() *Scene {
	// 寻找人数最多且小于最大人数限制的房间
	var maxScene *Scene
	maxPlayers := 0

	for _, sceneID := range sm.sceneJoinOrder {
		room := sm.scenes[sceneID]
		if room != nil {
			playerCount := room.GetPlayerCount()
			if playerCount < room.MaxPlayerNum && playerCount >= maxPlayers {
				maxPlayers = playerCount
				maxScene = room
			}
		}
	}

	// 如果没有可用房间，创建一个新房间
	if maxScene == nil {
		return sm.createScene()
	}

	return maxScene
}

func (sm *SceneManager) createScene() *Scene {
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
	scene.init(sceneID, sm.sceneType, sm.sceneMaxPlayerNum, sm.enableAOI, sm.aoiDistance)

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

func (sm *SceneManager) syncAOIInfoTicker() {
	go func() {
		ticker := time.Tick(sm.tickRate)
		for {
			select {
			case <-ticker:
				sm.mutex.RLock()
				for _, scene := range sm.GetScenes() {
					//log.Infof("%s 当前人数:%d", scene, scene.GetPlayerCount())
					scene.mutex.RLock()
					for _, player := range scene.players {
						submitSceneTask(func() {
							player.CollectAOISyncInfos()
						})
					}
					scene.mutex.RUnlock()
				}
				sm.mutex.RUnlock()
			}
		}

	}()
}

func submitSceneTask(task func()) {
	sceneMsgChan <- task
}

func GetSceneMsgChannelSize() int {
	return len(sceneMsgChan)
}

// RegisterSceneType register a entity type
func RegisterSceneType(scene IScene) {
	if sceneTypes[scene.GetSceneType()] != nil {
		log.Panic("scene type already registered, sceneType:" + scene.GetSceneType())
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
		log.Panic("scene type not found, sceneType:" + sceneType)
	}

	return objType
}

func GetSceneManager(sceneType string) *SceneManager {
	sceneManager := SceneManagerContext[sceneType]

	if sceneManager == nil {
		log.Warnf("scene manager not found, sceneType:%s", sceneType)
	}

	return sceneManager
}
