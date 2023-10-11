package gate_manager

import (
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/robfig/cron/v3"
	"go-one/common/consts"
	"go-one/common/log"
	"go-one/common/register"
	"sort"
	"strconv"
	"sync"
	"time"
)

var gateContext = make(map[int64]*GateInfos)
var gatesMutex = new(sync.RWMutex)

var crontab = *cron.New(cron.WithSeconds())

func init() {
	_, err := crontab.AddFunc("@every 10s", func() {
		start := time.Now().UnixMilli()
		FreshGateInfo()
		gatesMutex.RLock()
		for partition, gateInfos := range gateContext {
			gatesMutex.RUnlock()
			for _, gateInfo := range gateInfos.Gates {
				log.Infof("partition: %d, clusterId: %d, wsAddr: %s, tcpAddr: %s, version: %s, status: %d, connectionCount: %d",
					partition, gateInfo.ClusterId, gateInfo.WsAddr, gateInfo.TcpAddr, gateInfo.Version, gateInfo.Status, gateInfo.ConnectionCount)
			}
			gatesMutex.RLock()
		}
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

type GateInfo struct {
	Partition       int64
	ClusterId       int64
	WsAddr          string
	TcpAddr         string
	Version         string
	Status          int64
	ConnectionCount int64
}

func GetGateInfos(partition int64) *GateInfos {
	gatesMutex.RLock()
	gateInfos := gateContext[partition]
	gatesMutex.RUnlock()
	if gateInfos == nil {
		gatesMutex.Lock()
		gateInfos = gateContext[partition]
		if gateInfos != nil {
			gatesMutex.Unlock()
			return gateInfos
		}
		gateInfos = &GateInfos{
			Gates:      make(map[int64]*GateInfo),
			ClusterIds: make([]int64, 0),
		}
		gateContext[partition] = gateInfos
		gatesMutex.Unlock()
	}

	return gateInfos
}

func FreshGateInfo() {
	instances, err := register.NacosClient.SelectInstances(vo.SelectInstancesParam{
		ServiceName: "gate",
		HealthyOnly: true,
	})

	if err != nil {
		log.Warnf("select gate server instances error: %s", err.Error())
		return
	}

	if len(instances) == 0 {
		log.Warn("select gate server instances is empty")
		return
	}

	var instancesMap = make(map[int64]map[int64]bool)

	for _, instance := range instances {
		partitionStr := instance.Metadata[consts.Partition]
		clusterIdStr := instance.Metadata[consts.ClusterId]
		wsAddr := instance.Metadata[consts.WSAddr]
		tcpAddr := instance.Metadata[consts.TCPAddr]
		version := instance.Metadata[consts.Version]
		statusStr := instance.Metadata[consts.Status]

		partition, err := strconv.ParseInt(partitionStr, 10, 64)
		if err != nil {
			log.Errorf("gate partition is not int: %s", partitionStr)
			continue
		}

		clusterId, err := strconv.ParseInt(clusterIdStr, 10, 64)
		if err != nil {
			log.Errorf("gate clusterId is not int: %s", clusterIdStr)
			continue
		}

		status, err := strconv.ParseInt(statusStr, 10, 64)
		if err != nil {
			log.Warnf("gate status is not int: %s", clusterIdStr)
		}

		if status != consts.ServiceOnline {
			continue
		}

		instanceClusterIds := instancesMap[partition]
		if instanceClusterIds == nil {
			instanceClusterIds = make(map[int64]bool)
			instancesMap[partition] = instanceClusterIds
		}
		instanceClusterIds[clusterId] = true

		gateInfos := GetGateInfos(partition)

		gateInfo := gateInfos.getGate(clusterId)
		if gateInfo == nil {
			gateInfo = &GateInfo{
				Partition: partition,
				ClusterId: clusterId,
				WsAddr:    wsAddr,
				TcpAddr:   tcpAddr,
				Version:   version,
				Status:    status,
			}
			gateInfos.addGate(gateInfo)
		} else {
			gateInfo.WsAddr = wsAddr
			gateInfo.TcpAddr = tcpAddr
			gateInfo.Version = version
			gateInfo.Status = status
		}
	}

	for partition, clusterIds := range instancesMap {
		gateInfos := GetGateInfos(partition)
		gateInfos.Lock()
		for _, info := range gateInfos.Gates {
			if clusterIds[info.ClusterId] == false {
				delete(gateInfos.Gates, info.ClusterId)
			}
			var newClusterIds = make([]int64, 0)
			for clusterID, _ := range gateInfos.Gates {
				newClusterIds = append(newClusterIds, clusterID)
			}
			sort.Slice(newClusterIds, func(i, j int) bool {
				return newClusterIds[i] < newClusterIds[j]
			})
			gateInfos.ClusterIds = newClusterIds
		}
		gateInfos.Unlock()
	}
}

func ChooseGateInfo(partition int64, entityID int64) *GateInfo {
	gateInfos := GetGateInfos(partition)

	index := entityID % int64(len(gateInfos.ClusterIds))
	return gateInfos.getGate(gateInfos.ClusterIds[index])
}

func GetGateInfo(partition int64, clusterID int64) *GateInfo {
	gateInfos := GetGateInfos(partition)
	if gateInfos == nil {
		return nil
	}
	return gateInfos.getGate(clusterID)
}
