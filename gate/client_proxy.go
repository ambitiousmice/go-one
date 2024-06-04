package gate

import (
	context2 "context"
	"fmt"
	"github.com/ambitiousmice/go-one/common/common_proto"
	"github.com/ambitiousmice/go-one/common/consts"
	"github.com/ambitiousmice/go-one/common/context"
	"github.com/ambitiousmice/go-one/common/log"
	"github.com/ambitiousmice/go-one/common/pktconn"
	"github.com/ambitiousmice/go-one/gate/dispatcher"
	"github.com/ambitiousmice/go-one/gate/mq/kafka"
	"github.com/robfig/cron/v3"
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
	region    int32

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
		cp.SendCommonError(consts.GameMaintenance)
		return
	}

	err := gameDispatcher.ForwardMsg(cp.entityID, packet)
	if err != nil {
		log.Errorf("gameDispatcher.ForwardMsg error: %s", err)
		cp.SendCommonError(consts.GameMaintenance)
	}
}

// ============================================================================游戏协议============================================================================

func (cp *ClientProxy) Login(packet *pktconn.Packet) {
	log.Infof("%s start login", cp)
	if cp.entityID != 0 {
		log.Warnf("ready enter game, but already enter game: %s", cp)
		cp.SendEnterGameClientAck()
		cp.NotifyNewPlayerConnection()
		return
	}

	var param common_proto.LoginReq
	packet.ReadData(&param)

	if param.EntityID != 0 && param.ClientID != "" {
		oldCP := gateServer.getClientProxy(param.EntityID)
		if oldCP != nil && oldCP.entityID == param.EntityID && oldCP.clientID == param.ClientID {
			oldCP.CloseAll()
			cp.game = param.Game
			cp.clusterID = oldCP.clusterID
			cp.entityID = oldCP.entityID
		} else {
			log.Errorf("Reconnection failed: %s", cp)
			cp.SendError(common_proto.LoginFromClient, consts.ReconnectionFailed)
			return
		}

		cp.removeCronTask(consts.CheckLogin)

		gateServer.removeTempClientProxy(cp.clientID)

		gateServer.addClientProxy(cp)

		cp.NotifyNewPlayerConnection()

		log.Infof("reconnection success: %s", cp)

		return
	}

	loginResult, err := Login(gateServer.LoginManager, &param)
	if err != nil || !loginResult.Success {
		log.Errorf("Login error: %s", err)
		cp.SendError(common_proto.LoginFromClient, consts.LoginFailed)
		return
	}

	cp.entityID = loginResult.EntityID
	//

	oldCP := gateServer.getClientProxy(cp.entityID)
	if oldCP != nil {
		cp.clusterID = oldCP.clusterID
		oldCP.CloseAll()
	}

	kafka.SendGateLoginSyncNotify(cp.entityID, cp.clientID)
	cp.game = param.Game

	cp.region = loginResult.Region

	cp.removeCronTask(consts.CheckLogin)

	gateServer.removeTempClientProxy(cp.clientID)

	gateServer.addClientProxy(cp)

	cp.NotifyNewPlayerConnection()

	cp.SendEnterGameClientAck()

	log.Infof("login game success: %s", cp)
}

func (cp *ClientProxy) NotifyNewPlayerConnection() {
	packet := pktconn.NewPacket()
	packet.WriteUint16(common_proto.NewPlayerConnectionFromDispatcher)

	req := &common_proto.NewPlayerConnectionReq{
		EntityID: cp.entityID,
		Region:   cp.region,
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

	req := &common_proto.PlayerDisconnectedReq{
		EntityID: cp.entityID,
	}

	packet.AppendData(req)

	cp.ForwardByDispatcher(packet)

	cp.SendOffline()

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
	packet.WriteInt32(0)

	if msg != nil {
		packet.AppendData(msg)
	} else {
		packet.WriteUint32(0)
	}

	cp.SendAndRelease(packet)
}

func (cp *ClientProxy) SendCommonError(errorCode int32) {
	cp.SendError(common_proto.Error, errorCode)
}

func (cp *ClientProxy) SendError(cmd uint16, errorCode int32) {
	log.Warnf("%s send error ,cmd: %d code:%s", cp, cmd, errorCode)
	packet := pktconn.NewPacket()
	packet.WriteUint16(cmd)
	packet.WriteInt32(errorCode)

	packet.WriteUint32(0)

	cp.SendAndRelease(packet)
}

func (cp *ClientProxy) SendConnectionSuccessFromServer() {
	cp.SendMsg(common_proto.ConnectionSuccessFromServer, &common_proto.ConnectionSuccessFromServerResp{
		ClientID: cp.clientID,
	})
}

func (cp *ClientProxy) SendHeartBeatAck() {
	cp.SendMsg(common_proto.HeartbeatFromClientAck, &common_proto.HeartbeatAck{
		Time: time.Now().UnixMilli(),
	})
}

func (cp *ClientProxy) SendOffline() {
	cp.SendMsg(common_proto.OfflineFromServer, nil)
}

func (cp *ClientProxy) SendEnterGameClientAck() {
	cp.SendMsg(common_proto.LoginFromClientAck, &common_proto.LoginResp{
		EntityID: cp.entityID,
		Game:     cp.game,
	})
}
