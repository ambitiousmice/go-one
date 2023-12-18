package game_client

import (
	"go-one/common/common_proto"
	"go-one/common/log"
)

type AOISyncProcessor struct {
}

func (p *AOISyncProcessor) Process(client *Client, param []byte) {
	log.Infof("sync info:%s", param)

	var syncInfoBatch common_proto.AOISyncInfoBatch
	UnPackMsg(param, &syncInfoBatch)

	log.Infof("syncInfos size:%d", len(syncInfoBatch.GetSyncInfos()))
	/*for _, info := range syncInfos {
		s, _ := json.MarshalToString(info)
		log.Infof("sync info:%s", s)
	}*/
}

func (p *AOISyncProcessor) GetCmd() uint16 {
	return common_proto.AOISync
}
