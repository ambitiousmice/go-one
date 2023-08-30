package gate

import (
	context2 "context"
	"fmt"
	"github.com/robfig/cron/v3"
	"go-one/common/consts"
	"go-one/common/context"
	"go-one/common/log"
	"go-one/common/pktconn"
	"go-one/common/proto"
	"go-one/gate/dispatcher"
	"net"
	"time"
)

// ClientProxy is a game client connections managed by gate
type ClientProxy struct {
	*pktconn.PacketConn
	clientID string
	entityID int64
	game     string
	gameID   uint8

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
		} else {
			log.Debugf("%s disconnected", cp)
		}
	}()

	err := cp.ReceiveChan(gateServer.clientPacketQueue)
	if err != nil {
		log.Panic(err)
	}
}

func (cp *ClientProxy) CloseAll() {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("%s CloseAll error: %s", cp, err.(error))
		}
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
	if cp.gameID != 0 {
		gameDispatcher = dispatcher.GetGameDispatcher(cp.game, cp.gameID)
	} else {
		gameDispatcher = dispatcher.ChooseGameDispatcher(cp.game, cp.entityID)
		cp.gameID = gameDispatcher.GetGameID()
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

var ID int64 = 1

func (cp *ClientProxy) Login(packet *pktconn.Packet) {
	if cp.entityID != 0 {
		log.Warnf("ready login, but already login: %s", cp)
		cp.SendLoginAck()
		return
	}

	var param proto.LoginReq
	packet.ReadData(&param)

	/*loginResult, err := Login(gateServer.LoginManager, param)
	if err != nil {
		log.Errorf("Login error: %s", err)
		return
	}

	cp.entityID = loginResult.EntityID*/
	ID = ID + 1
	cp.entityID = ID
	cp.game = param.Game

	cp.cron.Remove(cp.cronMap[consts.CheckLogin])
	delete(cp.cronMap, consts.CheckLogin)

	cp.NewPlayerConnection()

	cp.SendLoginAck()

	log.Infof("Login success: %s", cp)
}

func (cp *ClientProxy) NewPlayerConnection() {
	packet := pktconn.NewPacket()
	packet.WriteUint16(proto.NewPlayerConnectionFromDispatcher)

	req := proto.NewPlayerConnectionReq{
		ClientID: cp.clientID,
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
	if cp.gameID == 0 {
		return
	}
	packet := pktconn.NewPacket()
	packet.WriteUint16(proto.PlayerDisconnectedFromDispatcher)

	req := proto.PlayerDisconnectedReq{
		ClientID: cp.clientID,
		EntityID: cp.entityID,
	}

	packet.AppendData(req)

	cp.ForwardByDispatcher(packet)

}

// ============================================================================客户端协议============================================================================

func (cp *ClientProxy) SendMsg(msgType uint16, msg interface{}) {
	packet := pktconn.NewPacket()
	packet.WriteUint16(msgType)

	if msg != nil {
		packet.AppendData(msg)
	}

	cp.SendAndRelease(packet)
}

func (cp *ClientProxy) SendError(error string) {
	cp.SendMsg(proto.Error, &proto.ErrorResp{
		Msg: error,
	})
}

func (cp *ClientProxy) SendNeedLoginFromServer() {
	cp.SendMsg(proto.NeedLoginFromServer, nil)
}

func (cp *ClientProxy) SendHeartBeatAck() {
	cp.SendMsg(proto.HeartbeatFromClientAck, time.Now().UnixMilli())
}

func (cp *ClientProxy) SendOffline() {
	cp.SendMsg(proto.OfflineFromServer, time.Now().UnixMilli())
}

func (cp *ClientProxy) SendLoginAck() {
	cp.SendMsg(proto.LoginFromClientAck, proto.LoginResp{
		EntityID: cp.entityID,
	})
}
