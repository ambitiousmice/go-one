package gate

import (
	context2 "context"
	"fmt"
	"github.com/ambitiousmice/go-one/common/common_proto"
	"github.com/ambitiousmice/go-one/common/consts"
	"github.com/ambitiousmice/go-one/common/context"
	"github.com/ambitiousmice/go-one/common/cust_error"
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

	status        uint8
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

func (cp *ClientProxy) ForwardByDispatcher(packet *pktconn.Packet) error {
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
		//cp.SendError("游戏维护中...")
		return cust_error.NewCustomError(common_proto.Game_Maintenance_Error, "游戏维护中...")
	}

	err := gameDispatcher.ForwardMsg(cp.entityID, packet)
	if err != nil {
		log.Errorf("gameDispatcher.ForwardMsg error: %s", err)
		//cp.SendError("游戏维护中...")
		return cust_error.NewCustomError(common_proto.Game_Maintenance_Error, "游戏维护中...")
	}
	return nil
}

// ============================================================================游戏协议============================================================================

func (cp *ClientProxy) Login(packet *pktconn.Packet) {
	log.Infof("%s start login", cp)
	if cp.entityID != 0 {
		log.Warnf("ready enter game, but already enter game: %s", cp)
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
			//cp.SendError("Reconnection failed")
			cp.SendMsg(common_proto.LoginFromClientAck, &common_proto.LoginResp{
				Code: common_proto.Game_Reconnect_Error,
			})
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
		if err != nil {
			customErr, ok := err.(*cust_error.CustomError)
			if ok {
				cp.SendMsg(common_proto.LoginFromClientAck, &common_proto.LoginResp{
					Code: customErr.ErrorCode,
				})
				return
			}
		}
		log.Errorf("Login error: %s", err)
		cp.SendMsg(common_proto.LoginFromClientAck, &common_proto.LoginResp{
			Code: common_proto.Game_Login_Error,
		})
		//cp.SendError("登录失败")
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

	err := cp.ForwardByDispatcher(packet)
	if err != nil {
		customErr, ok := err.(*cust_error.CustomError)
		if ok {
			cp.SendMsg(common_proto.LoginFromClientAck, &common_proto.LoginResp{
				Code: customErr.ErrorCode,
			})
		} else {
			cp.SendMsg(common_proto.LoginFromClientAck, &common_proto.LoginResp{
				Code: common_proto.Game_Maintenance_Error,
			})
		}
		return
	}
	cp.SendEnterGameClientAck()

}

func (cp *ClientProxy) HeartbeatTimeout() {
	log.Infof("客户端:%s 心跳超时", cp)

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

	if msg != nil {
		packet.AppendData(msg)
	}

	cp.SendAndRelease(packet)
}

func (cp *ClientProxy) SendError(error string) {
	log.Warnf("%s send error data:%s", cp, error)
	cp.SendMsg(common_proto.Error, &common_proto.ErrorResp{
		Data: error,
	})
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
	log.Infof("%s send offline", cp)
	cp.SendMsg(common_proto.OfflineFromServer, nil)
}

func (cp *ClientProxy) SendEnterGameClientAck() {
	cp.SendMsg(common_proto.LoginFromClientAck, &common_proto.LoginResp{
		EntityID: cp.entityID,
		Game:     cp.game,
	})
}
