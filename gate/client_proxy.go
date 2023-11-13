package gate

import (
	context2 "context"
	"fmt"
	"github.com/robfig/cron/v3"
	"go-one/common/common_proto"
	"go-one/common/consts"
	"go-one/common/context"
	"go-one/common/log"
	"go-one/common/pktconn"
	"go-one/gate/dispatcher"
	"net"
	"time"
)

// ClientProxy is a game client connections managed by gate
type ClientProxy struct {
	*pktconn.PacketConn
	clientID  string
	entityID  int64
	game      string
	clusterID uint8

	heartbeatTime time.Time
	cron          *cron.Cron
	cronMap       map[string]cron.EntryID
}

func newClientProxy(_conn net.Conn) *ClientProxy {

	_conn = pktconn.NewBufferedConn(_conn, consts.BufferedReadBufferSize, consts.BufferedWriteBufferSize)
	clientProxy := &ClientProxy{
		clientID:      context.NextClientID(),
		heartbeatTime: time.Now(),
		cron:          cron.New(cron.WithSeconds()),
		cronMap:       map[string]cron.EntryID{},
	}

	clientProxy.cron.Start()

	clientProxy.PacketConn = pktconn.NewPacketConn(context2.Background(), _conn, clientProxy)
	return clientProxy
}

func (cp *ClientProxy) String() string {
	return fmt.Sprintf("ClientProxy<%s@%d@%s>", cp.clientID, cp.entityID, cp.RemoteAddr())
}

func (cp *ClientProxy) serve() {
	defer func() {
		cp.CloseAll()
		if err := recover(); err != nil {
			log.Errorf("%s error: %s", cp, err.(error))
		}
	}()

	err := cp.ReceiveChan(gateServer.clientPacketQueue)
	if err != nil {
		log.Error(err)
	}
}

func (cp *ClientProxy) CloseAll() {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("%s CloseAll error: %s", cp, err.(error))
		}
		log.Infof("%s disconnected...", cp)
	}()

	cp.cron.Stop()

	cp.PlayerDisconnected()

	gateServer.onClientProxyClose(cp)

	cp.Close()

}

func (cp *ClientProxy) ForwardByDispatcher(packet *pktconn.Packet) {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("%s sendByDispatcher error: %s", cp, err.(error))
		}
	}()

	var gameDispatcher *dispatcher.GameDispatcher
	if cp.clusterID != 0 {
		gameDispatcher = dispatcher.GetGameDispatcher(cp.game, cp.clusterID)
		if gameDispatcher == nil {
			gameDispatcher = dispatcher.ChooseGameDispatcher(cp.game, cp.entityID)
			if gameDispatcher != nil {
				cp.clusterID = gameDispatcher.GetGameClusterID()
			}
		}
	} else {
		gameDispatcher = dispatcher.ChooseGameDispatcher(cp.game, cp.entityID)
		if gameDispatcher != nil {
			cp.clusterID = gameDispatcher.GetGameClusterID()
		}
	}

	if gameDispatcher == nil {
		log.Errorf("gameDispatcher is nil: %s", cp)
		cp.SendError("游戏维护中...")
		return
	}

	err := gameDispatcher.ForwardMsg(cp.entityID, packet)
	if err != nil {
		log.Errorf("gameDispatcher.ForwardMsg error: %s", err)
		cp.SendError("游戏维护中...")
	}
}

// ============================================================================游戏协议============================================================================

func (cp *ClientProxy) EnterGame(packet *pktconn.Packet) {
	if cp.entityID != 0 {
		log.Warnf("ready enter game, but already enter game: %s", cp)
		cp.SendEnterGameClientAck()
		cp.NotifyNewPlayerConnection()
		return
	}

	var param common_proto.EnterGameReq
	packet.ReadData(&param)

	if param.EntityID != 0 && param.ClientID != "" {
		oldCP := gateServer.getClientProxy(param.EntityID)
		if oldCP != nil && oldCP.entityID == param.EntityID && oldCP.clientID == param.ClientID {
			oldCP.cron.Stop()
			oldCP.Close()
			cp.game = param.Game
			cp.clusterID = oldCP.clusterID
			cp.entityID = oldCP.entityID
		} else {
			log.Errorf("Reconnection failed: %s", cp)
			cp.SendError("Reconnection failed")
			return
		}

		cp.removeCronTask(consts.CheckEnterGame)

		gateServer.removeTempClientProxy(cp.clientID)

		gateServer.addClientProxy(cp)

		cp.NotifyNewPlayerConnection()

		log.Infof("reconnection success: %s", cp)

		return
	}

	loginResult, err := Login(gateServer.LoginManager, param)
	if err != nil {
		log.Errorf("EnterGame error: %s", err)
		return
	}
	cp.entityID = loginResult.EntityID

	cp.game = param.Game

	cp.removeCronTask(consts.CheckEnterGame)

	gateServer.removeTempClientProxy(cp.clientID)

	gateServer.addClientProxy(cp)

	cp.NotifyNewPlayerConnection()

	cp.SendEnterGameClientAck()

	log.Infof("enter game success: %s", cp)
}

func (cp *ClientProxy) NotifyNewPlayerConnection() {
	packet := pktconn.NewPacket()
	packet.WriteUint16(common_proto.NewPlayerConnectionFromDispatcher)

	req := common_proto.NewPlayerConnectionReq{
		EntityID: cp.entityID,
	}

	packet.AppendData(req)

	cp.ForwardByDispatcher(packet)

}

func (cp *ClientProxy) HeartbeatTimeout() {
	log.Infof("客户端:%s 超时", cp)

	cp.CloseAll()
}

func (cp *ClientProxy) PlayerDisconnected() {
	if cp.clusterID == 0 || cp.entityID == 0 {
		return
	}
	packet := pktconn.NewPacket()
	packet.WriteUint16(common_proto.PlayerDisconnectedFromDispatcher)

	req := common_proto.PlayerDisconnectedReq{
		EntityID: cp.entityID,
	}

	packet.AppendData(req)

	cp.ForwardByDispatcher(packet)

}

func (cp *ClientProxy) removeCronTask(taskName string) {
	taskID, ok := cp.cronMap[taskName]
	if !ok {
		return
	}

	cp.cron.Remove(taskID)
	delete(cp.cronMap, taskName)
}

// ============================================================================客户端协议============================================================================

func (cp *ClientProxy) SendMsg(msgType uint16, msg any) {
	packet := pktconn.NewPacket()
	packet.WriteUint16(msgType)

	if msg != nil {
		packet.AppendData(msg)
	}

	cp.SendAndRelease(packet)
}

func (cp *ClientProxy) SendError(error string) {
	cp.SendMsg(common_proto.Error, &common_proto.ErrorResp{
		Data: error,
	})
}

func (cp *ClientProxy) SendConnectionSuccessFromServer() {
	cp.SendMsg(common_proto.ConnectionSuccessFromServer, &common_proto.EnterGameFromServerParam{
		ClientID: cp.clientID,
	})
}

func (cp *ClientProxy) SendHeartBeatAck() {
	cp.SendMsg(common_proto.HeartbeatFromClientAck, time.Now().UnixMilli())
}

func (cp *ClientProxy) SendOffline() {
	cp.SendMsg(common_proto.OfflineFromServer, time.Now().UnixMilli())
}

func (cp *ClientProxy) SendEnterGameClientAck() {
	cp.SendMsg(common_proto.EnterGameClientAck, common_proto.EnterGameResp{
		EntityID: cp.entityID,
		Game:     cp.game,
	})
}
