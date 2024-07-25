package server_manager

import (
	"github.com/ambitiousmice/go-one/common/cache"
	"github.com/ambitiousmice/go-one/common/json"
	"github.com/ambitiousmice/go-one/common/log"
	"github.com/ambitiousmice/go-one/common/utils"
	"github.com/ambitiousmice/go-one/monitor/config"
	"github.com/robfig/cron/v3"
	"sort"
	"strconv"
	"time"
)

var crontab = *cron.New(cron.WithSeconds())

var entityGateInfoCacheKey = "entity_gate_map"
var serverInfosCacheKey = "m_server_infos"

var serverContext = make(map[string]*ServerInfo)

var gateContext = make(map[int64]*GateInfos) // gateContext: {groupID: {gates: []*GateInfo}}
var regionGroupMap = make(map[int64]int64)

func Init() {
	InitRegionClusterMap()
	_, err := crontab.AddFunc("@every 5s", func() {
		start := time.Now().UnixMilli()
		FreshServerInfo()
		log.Infof("fresh server info success, cost: %d ms", time.Now().UnixMilli()-start)
	})
	if err != nil {
		log.Panicf("init server manager crontab error: ", err.Error())
	}

	crontab.Start()
}

func AddServer(serverInfo *ServerInfo) {
	err := cache.SetHashField(serverInfosCacheKey, serverInfo.ServerName+"_"+utils.ToString(serverInfo.GroupID)+"_"+utils.ToString(serverInfo.ClusterId), serverInfo)
	if err != nil {
		log.Errorf("set server info error:%s", err)
	}
}

type ServerInfo struct {
	ServerName            string
	GroupID               int64
	ClusterId             int64
	ConnectionCount       int64
	LastCommunicationTime int64
	TotalMemory           float64
	UsageMemory           float64
	Status                int8
	Metadata              map[string]string
}

type GateInfos struct {
	Gates      map[int64]*ServerInfo
	ClusterIds []int64
}

func GetGateInfos(groupID int64) *GateInfos {
	return gateContext[groupID]
}

func FreshServerInfo() {
	servers, err := cache.GetHashAll(serverInfosCacheKey)

	if err != nil {
		log.Warnf("fresh server info failed,get cache error:%s", err)
		return
	}

	var tempServerMap = make(map[string]*ServerInfo)
	var tempGateContext = make(map[int64]*GateInfos)
	for _, v := range servers {
		var serverInfo ServerInfo
		json.UnmarshalFromString(v, &serverInfo)

		communicationTimeValue := time.Now().UnixMilli() - serverInfo.LastCommunicationTime
		communicationTimeoutValue := config.GetConfig().Gate.CommunicationTimeoutValue
		if communicationTimeoutValue == 0 {
			communicationTimeoutValue = 5
		}
		if communicationTimeValue > 1000*communicationTimeoutValue {
			log.Warnf("server communication time is too long:%s,%d,%d", serverInfo.ServerName, serverInfo.GroupID, serverInfo.ClusterId)
			cache.DeleteHashField(serverInfosCacheKey, serverInfo.ServerName+"_"+utils.ToString(serverInfo.GroupID)+"_"+utils.ToString(serverInfo.ClusterId))
			continue
		}
		tempServerMap[serverInfo.ServerName+"_"+utils.ToString(serverInfo.GroupID)+"_"+utils.ToString(serverInfo.ClusterId)] = &serverInfo

		serverInfoStr, _ := json.MarshalToString(serverInfo)
		log.Infof("%s 服务信息:%s", serverInfo.ServerName, serverInfoStr)

		if serverInfo.ServerName != config.GetConfig().Gate.Name {
			continue
		}

		gateInfos := tempGateContext[serverInfo.GroupID]
		if gateInfos == nil {
			gateInfos = &GateInfos{
				Gates:      make(map[int64]*ServerInfo),
				ClusterIds: make([]int64, 0),
			}
			tempGateContext[serverInfo.GroupID] = gateInfos
		}

		gateInfos.Gates[serverInfo.ClusterId] = &serverInfo
	}

	for _, infos := range tempGateContext {
		var newClusterIds = make([]int64, 0)
		for clusterID, _ := range infos.Gates {
			newClusterIds = append(newClusterIds, clusterID)
		}
		sort.Slice(newClusterIds, func(i, j int) bool {
			return newClusterIds[i] < newClusterIds[j]
		})
		infos.ClusterIds = newClusterIds
	}

	serverContext = tempServerMap
	gateContext = tempGateContext
}

func ChooseGateInfo(partition int64, entityID int64) *ServerInfo {
	groupId := regionGroupMap[partition]
	if groupId == 0 {
		log.Warnf("set entity gate cache error")
	}
	gateInfos := GetGateInfos(groupId)

	var previousGateInfo ServerInfo
	err := cache.GetHashField(entityGateInfoCacheKey, utils.ToString(entityID), &previousGateInfo)
	if err == nil && previousGateInfo.GroupID == groupId {
		newGateInfo := gateInfos.Gates[previousGateInfo.ClusterId]
		if newGateInfo != nil {
			return newGateInfo
		}
	}

	index := entityID % int64(len(gateInfos.ClusterIds))
	newGateInfo := gateInfos.Gates[gateInfos.ClusterIds[index]]

	err = cache.SetHashField(entityGateInfoCacheKey, utils.ToString(entityID), newGateInfo)
	if err != nil {
		log.Warnf("set entity gate cache error:%s", err.Error())
	}

	return newGateInfo
}

func GetGateInfo(groupID int64, clusterID int64) *ServerInfo {
	gateInfos := GetGateInfos(groupID)
	if gateInfos == nil {
		return nil
	}
	return gateInfos.Gates[clusterID]
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
