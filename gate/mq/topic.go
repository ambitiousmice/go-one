package mq

const (
	GateSyncPlayer = "GateSyncPlayer"
)

type GateLoginSyncNotify struct {
	EntityID    int64  `json:"entityID"`
	ClientID    string `json:"clientID"`
	Service     string `json:"service"`
	GateGroup   string `json:"gateGroup"`
	ClusterName string `json:"clusterName"`
}
