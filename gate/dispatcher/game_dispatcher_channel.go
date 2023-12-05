package dispatcher

import (
	context2 "context"
	"fmt"
	"github.com/robfig/cron/v3"
	"go-one/common/common_proto"
	"go-one/common/consts"
	"go-one/common/context"
	"go-one/common/log"
	"go-one/common/network"
	"go-one/common/pktconn"
	"go-one/common/utils"
	"net"
	"strconv"
	"sync"
	"time"
)

type GameDispatcherChannel struct {
	*pktconn.PacketConn
	gameDispatcher *GameDispatcher
	sync.RWMutex
	channelID uint8
	status    int8

	packetQueue                       chan *pktconn.Packet
	cron                              *cron.Cron
	ticker                            <-chan time.Time
	heartbeatTime                     time.Time
	tryReconnectedCount               uint8
	dispatcherClientPacketQueuesIndex uint64
}

func (gpc *GameDispatcherChannel) String() string {
	return fmt.Sprintf("GameDispatcherChannel<%s@%d@%s>", gpc.gameDispatcher.game, gpc.channelID, gpc.RemoteAddr())
}

func NewDispatcherChannel(channelID uint8, dispatcher *GameDispatcher) *GameDispatcherChannel {
	return &GameDispatcherChannel{
		channelID:      channelID,
		status:         consts.DispatcherChannelStatusInit,
		gameDispatcher: dispatcher,

		packetQueue:   make(chan *pktconn.Packet, consts.DispatcherChannelPacketQueueSize),
		cron:          cron.New(cron.WithSeconds()),
		ticker:        time.Tick(consts.ChannelTickInterval),
		heartbeatTime: time.Now(),
	}
}

func (gpc *GameDispatcherChannel) Run() {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("Game Dispatcher Channel Run, panic with error %s", err)
		}
	}()

	var netConn net.Conn
	netConn, err := gpc.connectServer()
	if err != nil {
		log.Errorf("connect server failed: " + err.Error())
		return
	}

	netConn = pktconn.NewBufferedConn(netConn, consts.BufferedReadBufferSize, consts.BufferedWriteBufferSize)
	gpc.PacketConn = pktconn.NewPacketConn(context2.Background(), netConn, gpc)

	gpc.cron.AddFunc("@every 2s", func() {
		gpc.sendHeartbeat()
	})

	gpc.cron.AddFunc("@every 5s", func() {
		if time.Now().Sub(gpc.heartbeatTime) > 5*time.Second {
			log.Infof("dispatcher channel status: %d", gpc.status)
			if gpc.status == consts.DispatcherChannelStatusUnHealth {
				return
			}

			if gpc.status == consts.DispatcherChannelStatusStop || gpc.status == consts.DispatcherChannelStatusRestart {
				return
			}

			gpc.updateStatus(consts.DispatcherChannelStatusUnHealth)
			log.Infof("%s heartbeat timeout, updating status to unhealthy ...", gpc)

		}
	})

	gpc.cron.Start()

	gpc.sendDispatcherInfo()

	go gpc.receive()

	gpc.updateStatus(consts.DispatcherChannelStatusHealth)

	log.Infof("game<%s> dispatcher channel<%d> connect to server: %s", gpc.gameDispatcher.game, gpc.channelID, netConn.RemoteAddr().String())

	gpc.handlePacketQueue()

}

func (gpc *GameDispatcherChannel) ReRun() {
	gpc.Lock()
	defer gpc.Unlock()
	if gpc.status == consts.DispatcherChannelStatusRestart || gpc.status == consts.DispatcherChannelStatusRestartFailed {
		return
	}
	gpc.status = consts.DispatcherChannelStatusRestart

	if gpc.tryReconnectedCount >= consts.DispatcherChannelMaxTryReconnectedCount {
		log.Infof("%s reconnection attempts has reached the maximum limit...", gpc)
		gpc.status = consts.DispatcherChannelStatusRestartFailed
		return
	}

	gpc.tryReconnectedCount++

	log.Infof("%s try reconnecting for the %d time...", gpc, gpc.tryReconnectedCount)
	netConn, err := gpc.connectServer()
	if err != nil {
		log.Errorf("%s reconnect server failed: %s", gpc, err.Error())
		gpc.status = consts.DispatcherChannelStatusUnHealth
		return
	}

	netConn = pktconn.NewBufferedConn(netConn, consts.BufferedReadBufferSize, consts.BufferedWriteBufferSize)
	gpc.PacketConn = pktconn.NewPacketConn(context2.Background(), netConn, gpc)
	gpc.heartbeatTime = time.Now()

	gpc.sendDispatcherInfo()

	go gpc.receive()

	gpc.status = consts.DispatcherChannelStatusHealth

	log.Infof("game<%s> dispatcher channel<%d> reconnect to server: %s", gpc.gameDispatcher.game, gpc.channelID, gpc.RemoteAddr().String())
}

func (gpc *GameDispatcherChannel) stop() {
	gpc.updateStatus(consts.DispatcherChannelStatusStop)
	gpc.cron.Stop()
	gpc.Close()
	close(gpc.packetQueue)

	log.Warnf("game<%s> dispatcher channel<%d> closed: %s", gpc.gameDispatcher.game, gpc.channelID, gpc.RemoteAddr().String())

	gpc.gameDispatcher = nil
}

func (gpc *GameDispatcherChannel) connectServer() (net.Conn, error) {

	conn, err := network.ConnectTCP(net.JoinHostPort(gpc.gameDispatcher.gameHost, utils.ToString(gpc.gameDispatcher.gamePort)))
	if err == nil {
		conn.(*net.TCPConn).SetWriteBuffer(consts.GameDispatcherWriteBufferSize)
		conn.(*net.TCPConn).SetReadBuffer(consts.GameDispatcherReadBufferSize)
	}
	return conn, err
}

func (gpc *GameDispatcherChannel) receive() {
	err := gpc.ReceiveChan(gpc.packetQueue)
	if err != nil {
		log.Error(err)
	}

}

func (gpc *GameDispatcherChannel) handlePacketQueue() {
	for {
		select {
		case pkt, ok := <-gpc.packetQueue:
			if !ok {
				return
			}
			gpc.handleGameMsg(pkt)
			pkt.Release()
		case <-gpc.ticker:
			if gpc.status == consts.DispatcherChannelStatusStop {
				log.Infof("%s handlePacketQueue exit...", gpc)
				return
			}
		}
	}
}

func (gpc *GameDispatcherChannel) updateHeartbeatTime() {
	gpc.heartbeatTime = time.Now()
}

func (gpc *GameDispatcherChannel) updateStatus(status int8) {
	gpc.Lock()
	defer gpc.Unlock()
	gpc.status = status
	log.Infof("%s updateStatus to %d", gpc, status)
}

func (gpc *GameDispatcherChannel) getStatus() int8 {
	gpc.RLock()
	defer gpc.RUnlock()
	return gpc.status
}

func (gpc *GameDispatcherChannel) handleGameMsg(packet *pktconn.Packet) {
	defer func() {
		err := recover()
		if err != nil {
			log.Errorf("handle packet failed: %v", err)
		}
	}()

	msgType := packet.ReadUint16()
	/*if msgType == common_proto.GameMethodFromClientAck {
		log.Infof("handleGameMsg: %d", msgType)
	}*/

	switch msgType {
	case common_proto.HeartbeatFromDispatcherAck:
		gpc.updateHeartbeatTime()
	case common_proto.GameMethodFromClientAck:
		gpc.processAck2000(packet)
	case common_proto.GameDispatcherChannelInfoFromDispatcherAck:
		Handle3002(packet)

	default:
		log.Errorf("unknown msgType: %d", msgType)
	}
}

func (gpc *GameDispatcherChannel) processAck2000(packet *pktconn.Packet) {
	packet.Retain()
	dispatcherClientPacketQueue := getDispatcherClientPacketQueue()
	select {
	case dispatcherClientPacketQueue <- packet:
		// 数据包成功发送到队列
	default:
		// 队列已满或其他原因，不阻塞发送
		packet.Release() // 释放多余的数据包
		log.Warnf("dispatcherClientPacketQueues is full, drop packet")
	}
}

// message handler======================================================================================================

func Handle3002(pkt *pktconn.Packet) {
	req := &common_proto.GameDispatcherChannelInfoResp{}
	pkt.ReadData(req)
	log.Infof("handle3002: %v", req)

	if !req.Success {
		log.Errorf("handle3002: %s", req.Msg)
		log.Panic("handle3002 failed")
	}
}

// =========================================================

func (gpc *GameDispatcherChannel) SendMsg(msgType uint16, msg interface{}) {
	packet := pktconn.NewPacket()
	packet.WriteUint16(msgType)
	if msg != nil {
		packet.AppendData(msg)
	}
	gpc.SendAndRelease(packet)
}

func (gpc *GameDispatcherChannel) sendHeartbeat() {
	gpc.SendMsg(common_proto.HeartbeatFromDispatcher, nil)
}

func (gpc *GameDispatcherChannel) sendDispatcherInfo() {
	clusterIDStr := context.GetOneConfig().Nacos.Instance.ClusterName

	clusterID, _ := strconv.ParseUint(clusterIDStr, 10, 8)
	gpc.SendMsg(common_proto.GameDispatcherChannelInfoFromDispatcher, &common_proto.GameDispatcherChannelInfoReq{
		GateClusterID: uint8(clusterID),
		Game:          gpc.gameDispatcher.game,
		GameClusterID: gpc.gameDispatcher.gameClusterID,
		ChannelID:     gpc.channelID,
	})
}
