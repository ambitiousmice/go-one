package game_client

import (
	"go-one/common/common_proto"
	"go-one/common/log"
)

type AOISyncProcessor struct {
}

func (p *AOISyncProcessor) Process(client *Client, param []byte) {
	//log.Infof("sync info:%s", param)

	var syncInfos []common_proto.AOISyncInfo
	UnPackMsg(param, &syncInfos)

	log.Infof("syncInfos size:%d", len(syncInfos))
	/*for _, info := range syncInfos {
		s, _ := json.MarshalToString(info)
		log.Infof("sync info:%s", s)
	}*/
}

func (p *AOISyncProcessor) GetCmd() uint16 {
	return common_proto.AOISync
}
