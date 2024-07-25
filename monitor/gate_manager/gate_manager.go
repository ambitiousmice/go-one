package gate_manager

import (
	"github.com/ambitiousmice/go-one/common/cache"
	"github.com/ambitiousmice/go-one/common/consts"
	"github.com/ambitiousmice/go-one/common/log"
	"github.com/ambitiousmice/go-one/common/register"
	"github.com/ambitiousmice/go-one/common/utils"
	"github.com/ambitiousmice/go-one/monitor/config"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/robfig/cron/v3"
	"sort"
	"strconv"
	"sync"
	"time"
)

// 已废弃

var gateContext = make(map[int64]*GateInfos) // gateContext: {groupID: {gates: []*GateInfo}}
var gatesMutex = new(sync.RWMutex)

var regionGroupMap = make(map[int64]int64) // regionGroupMap: {regionID: groupID}

var crontab = *cron.New(cron.WithSeconds())

var entityGateInfoCacheKey = "entity_gate_map"

func Init() {
	InitRegionClusterMap()
	_, err := crontab.AddFunc("@every 10s", func() {
		start := time.Now().UnixMilli()
		for groupName, _ := range config.GetConfig().Gate.GroupInfos {
			FreshGateInfo(config.GetConfig().Gate.Name, groupName)
		}
		gatesMutex.RLock()
		for groupID, gateInfos := range gateContext {
			gateInfos.RLock()
			for _, gateInfo := range gateInfos.Gates {
				log.Infof("groupID: %d, clusterId: %d, wsAddr: %s, tcpAddr: %s, version: %s, status: %d, connectionCount: %d",
					groupID, gateInfo.ClusterId, gateInfo.WsAddr, gateInfo.TcpAddr, gateInfo.Version, gateInfo.Status, gateInfo.ConnectionCount)
			}
			gateInfos.RUnlock()
		}
		gatesMutex.RUnlock()
		log.Infof("fresh gate info success, cost: %d ms", time.Now().UnixMilli()-start)
	})
	if err != nil {
		log.Panicf("init gate manager crontab error: ", err.Error())
	}

	crontab.Start()
}

type GateInfos struct {
	sync.RWMutex
	Gates      map[int64]*GateInfo
	ClusterIds []int64
}

func (g *GateInfos) getGate(clusterId int64) *GateInfo {
	g.RLock()
	defer g.RUnlock()
	gate := g.Gates[clusterId]
	return gate
}

func (g *GateInfos) addGate(gateInfo *GateInfo) {
	g.Lock()
	defer g.Unlock()
	g.Gates[gateInfo.ClusterId] = gateInfo
	g.ClusterIds = append(g.ClusterIds, gateInfo.ClusterId)
}

func (g *GateInfos) removeGate(clusterId int64) {
	g.Lock()
	defer g.Unlock()
	delete(g.Gates, clusterId)
}

type GateInfo struct {
	GroupID               int64
	ClusterId             int64
	WsAddr                string
	TcpAddr               string
	Version               string
	Status                int64
	ConnectionCount       int64
	LastCommunicationTime int64
}

func GetGateInfos(groupID int64) *GateInfos {
	gatesMutex.RLock()
	gateInfos := gateContext[groupID]
	gatesMutex.RUnlock()
	if gateInfos == nil {
		gatesMutex.Lock()
		gateInfos = gateContext[groupID]
		if gateInfos != nil {
			gatesMutex.Unlock()
			return gateInfos
		}
		gateInfos = &GateInfos{
			Gates:      make(map[int64]*GateInfo),
			ClusterIds: make([]int64, 0),
		}
		gateContext[groupID] = gateInfos
		gatesMutex.Unlock()
	}

	return gateInfos
}

func FreshGateInfo(gateName, groupName string) {
	instances, err := register.NacosClient.SelectInstances(vo.SelectInstancesParam{
		ServiceName: gateName,
		GroupName:   groupName,
		HealthyOnly: true,
	})

	if err != nil {
		log.Warnf("select %s|%s server error:%s", groupName, gateName, err.Error())
	}

	groupID, err := strconv.ParseInt(groupName, 10, 64)
	if err != nil {
		log.Errorf("gate groupID is not int: %s", groupName)
		return
	}

	var existingInstances = make(map[int64]bool)

	gateInfos := GetGateInfos(groupID)
	for _, info := range gateInfos.Gates {
		existingInstances[info.ClusterId] = true
	}

	for clusterID := range existingInstances {
		exist := false
		for _, instance := range instances {
			clusterIDStr := instance.Metadata[consts.ClusterId]
			clusterId, _ := strconv.ParseInt(clusterIDStr, 10, 64)
			if clusterId == clusterID {
				exist = true
				break
			}
		}
		if !exist {
			gateInfos.removeGate(clusterID)
		}
	}

	for _, instance := range instances {
		clusterIdStr := instance.Metadata[consts.ClusterId]
		clusterId, err := strconv.ParseInt(clusterIdStr, 10, 64)
		if err != nil {
			log.Errorf("gate clusterId is not int,instance info:%s:%s|%s", instance.Ip, instance.Port, clusterIdStr)
			continue
		}

		wsAddr := instance.Metadata[consts.WSAddr]
		tcpAddr := instance.Metadata[consts.TCPAddr]
		version := instance.Metadata[consts.Version]
		statusStr := instance.Metadata[consts.Status]

		status, err := strconv.ParseInt(statusStr, 10, 64)
		if err != nil {
			log.Warnf("gate status is not int: %s", clusterIdStr)
		}

		if status != consts.ServiceOnline {
			continue
		}

		gateInfo := gateInfos.getGate(clusterId)
		if gateInfo == nil {
			gateInfo = &GateInfo{
				GroupID:               groupID,
				ClusterId:             clusterId,
				WsAddr:                wsAddr,
				TcpAddr:               tcpAddr,
				Version:               version,
				Status:                status,
				LastCommunicationTime: time.Now().UnixMilli(),
			}
			gateInfos.addGate(gateInfo)
		} else {
			gateInfo.WsAddr = wsAddr
			gateInfo.TcpAddr = tcpAddr
			gateInfo.Version = version
			gateInfo.Status = status
			gateInfo.LastCommunicationTime = time.Now().UnixMilli()
		}
	}

	gateInfos.Lock()
	var newClusterIds = make([]int64, 0)
	for clusterID, _ := range gateInfos.Gates {
		newClusterIds = append(newClusterIds, clusterID)
	}
	sort.Slice(newClusterIds, func(i, j int) bool {
		return newClusterIds[i] < newClusterIds[j]
	})
	gateInfos.ClusterIds = newClusterIds
	gateInfos.Unlock()
}

func ChooseGateInfo(partition int64, entityID int64) *GateInfo {
	groupId := regionGroupMap[partition]
	if groupId == 0 {
		log.Warnf("set entity gate cache error")
	}
	gateInfos := GetGateInfos(groupId)

	var previousGateInfo GateInfo
	err := cache.GetHashField(entityGateInfoCacheKey, utils.ToString(entityID), &previousGateInfo)
	if err == nil && previousGateInfo.GroupID == groupId {
		newGateInfo := gateInfos.getGate(previousGateInfo.ClusterId)
		if newGateInfo != nil {
			return newGateInfo
		}
	}

	index := entityID % int64(len(gateInfos.ClusterIds))
	newGateInfo := gateInfos.getGate(gateInfos.ClusterIds[index])

	err = cache.SetHashField(entityGateInfoCacheKey, utils.ToString(entityID), newGateInfo)
	if err != nil {
		log.Warnf("set entity gate cache error:%s", err.Error())
	}

	return newGateInfo
}

func GetGateInfo(groupID int64, clusterID int64) *GateInfo {
	gateInfos := GetGateInfos(groupID)
	if gateInfos == nil {
		return nil
	}
	return gateInfos.getGate(clusterID)
}

func InitRegionClusterMap() {
	for groupName, regions := range config.GetConfig().Gate.GroupInfos {
		groupId, err := strconv.ParseInt(groupName, 10, 64)
		if err != nil {
			log.Panicf("groupName is not int: %s", groupName)
		}
		for _, regionStr := range regions {
			region, err := strconv.ParseInt(regionStr, 10, 64)
			if err != nil {
				log.Panicf("region is not int: %s", regionStr)
			}
			regionGroupMap[region] = groupId
		}
	}
}
