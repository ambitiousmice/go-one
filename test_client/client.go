package main

import (
	context2 "context"
	"github.com/ambitiousmice/go-one/common/common_proto"
	"github.com/ambitiousmice/go-one/common/consts"
	"github.com/ambitiousmice/go-one/common/json"
	"github.com/ambitiousmice/go-one/common/log"
	"github.com/ambitiousmice/go-one/common/network"
	"github.com/ambitiousmice/go-one/common/pktconn"
	"github.com/ambitiousmice/go-one/common/utils"
	"github.com/robfig/cron/v3"
	"net"
	"sync"

	"fmt"

	"time"

	"github.com/xtaci/kcp-go"
	"golang.org/x/net/websocket"
)

const _SPACE_ENTITY_TYPE = "__space__"

// ClientBot is  a client bot representing a game client
type ClientBot struct {
	sync.Mutex

	id          int
	ServerHost  string
	conn        *pktconn.PacketConn
	packetQueue chan *pktconn.Packet
	crontab     *cron.Cron
}

type result struct {
	code string
	data chooseGateResp
}
type chooseGateResp struct {
	WsAddr  string
	TcpAddr string
	Version string
}

func newClientBot(id int) *ClientBot {
	c := &ClientBot{
		id:          id,
		packetQueue: make(chan *pktconn.Packet),
		crontab:     cron.New(cron.WithSeconds()),
	}
	if Config.ServerConfig.UseLoadBalancer {
		param := make(map[string]string)
		param["partition"] = Config.ServerConfig.Partition
		param["userId"] = utils.ToString(id)

		resp, err := utils.Get(Config.ServerConfig.LoadBalancerUrl, param)
		if err != nil {
			log.Panic(err)
		}
		var r result
		json.UnmarshalFromString(resp, r)
		if Config.ServerConfig.Websocket {
			c.ServerHost = r.data.WsAddr
		} else {
			c.ServerHost = r.data.TcpAddr
		}
	} else {
		c.ServerHost = Config.ServerConfig.ServerHost
	}
	return c
}

func (bot *ClientBot) String() string {
	return fmt.Sprintf("ClientBot<%d>", bot.id)
}

func (bot *ClientBot) run() {

	var netConn net.Conn
	netConn, err := bot.connectServer()
	if err != nil {
		log.Panic("connect server failed: " + err.Error())
	}

	log.Infof("connected: %s", netConn.RemoteAddr())

	netConn = pktconn.NewBufferedConn(netConn, consts.BufferedReadBufferSize, consts.BufferedWriteBufferSize)
	bot.conn = pktconn.NewPacketConn(context2.Background(), netConn, bot)
	defer bot.conn.Close()
	bot.crontab.Start()

	bot.crontab.AddFunc("@every 2s", func() {
		packet := pktconn.NewPacket()
		packet.WriteUint16(common_proto.HeartbeatFromClient)
		bot.conn.SendAndRelease(packet)
		//log.Infof("==============发送心跳包,packetQueue长度:%d", len(bot.packetQueue))
	})
	// send handshake packet

	go bot.recvLoop()

	bot.loop()
}

func (bot *ClientBot) connectServer() (net.Conn, error) {
	if Config.ServerConfig.Websocket {
		return bot.connectServerByWebsocket()
	} else if Config.ServerConfig.Kcp {
		return bot.connectServerByKCP()
	}

	conn, err := network.ConnectTCP(bot.ServerHost)
	if err == nil {
		conn.(*net.TCPConn).SetWriteBuffer(64 * 1024)
		conn.(*net.TCPConn).SetReadBuffer(64 * 1024)
	}
	return conn, err
}

func (bot *ClientBot) connectServerByKCP() (net.Conn, error) {

	conn, err := kcp.DialWithOptions(bot.ServerHost, nil, 0, 0)
	if err != nil {
		return nil, err
	}
	conn.SetReadBuffer(64 * 1024)
	conn.SetWriteBuffer(64 * 1024)
	conn.SetNoDelay(consts.KCP_NO_DELAY, consts.KCP_INTERNAL_UPDATE_TIMER_INTERVAL, consts.KCP_ENABLE_FAST_RESEND, consts.KCP_DISABLE_CONGESTION_CONTROL)
	conn.SetStreamMode(consts.KCP_SET_STREAM_MODE)
	conn.SetWriteDelay(consts.KCP_SET_WRITE_DELAY)
	conn.SetACKNoDelay(consts.KCP_SET_ACK_NO_DELAY)
	return conn, err
}

func (bot *ClientBot) connectServerByWebsocket() (net.Conn, error) {
	originProto := "http"
	wsProto := "ws"

	origin := fmt.Sprintf("%s://%s/", originProto, bot.ServerHost)
	wsaddr := fmt.Sprintf("%s://%s/ws", wsProto, bot.ServerHost)

	return websocket.Dial(wsaddr, "", origin)
}

func (bot *ClientBot) recvLoop() {
	err := bot.conn.ReceiveChan(bot.packetQueue)
	log.Error(err)
}

func (bot *ClientBot) loop() {
	ticker := time.Tick(time.Millisecond * 100)
	for {
		select {
		case pkt := <-bot.packetQueue:
			bot.handlePacket(pkt)
			pkt.Release()
			//break
		case <-ticker:

			break
		}
	}
}

func (bot *ClientBot) handlePacket(packet *pktconn.Packet) {
	defer func() {
		err := recover()
		if err != nil {
			log.Errorf("handle packet failed: %v", err)
		}
	}()

	msgType := packet.ReadUint16()
	if msgType != 2001 {
		log.Infof("handlePacket: %d", msgType)
	}
	switch msgType {
	case common_proto.ConnectionSuccessFromServer:
		bot.login("190e5f8a-e3aa-4320-954d-8505b4393de4")
		log.Infof("发送登录消息")
	case common_proto.LoginFromClientAck:
		loginResp := &common_proto.LoginResp{}
		packet.ReadData(loginResp)
		log.Infof("登录结果,EntityID:%d", loginResp.EntityID)
		/*go func() {
			for true {
				bot.sendGameMsg(1, []byte("1"))
				time.Sleep(time.Microsecond * 100)
			}
		}()*/
	case common_proto.GameMethodFromClientAck:
		gameResp := &common_proto.GameResp{}
		packet.ReadData(gameResp)
		switch gameResp.Cmd {
		case common_proto.JoinScene:
			joinRoomResp := &common_proto.JoinSceneResp{}
			pktconn.MSG_PACKER.UnpackMsg(gameResp.Data, joinRoomResp)
			log.Infof("加入场景结果:%s", joinRoomResp)
		default:
			log.Infof("gameResp:%d,%s", gameResp.Cmd, string(gameResp.Data))

		}
		pktconn.MSG_PACKER.UnpackMsg(gameResp.Data, gameResp)
		log.Infof("gameResp:%d,%s", gameResp.Cmd, string(gameResp.Data))
	}

}

func (bot *ClientBot) SendMsg(msgType uint16, msg interface{}) {
	packet := pktconn.NewPacket()
	packet.WriteUint16(msgType)
	if msg != nil {
		packet.AppendData(msg)
	}
	bot.conn.Send(packet)
	packet.Release()
}

func (bot *ClientBot) login(account string) {
	bot.SendMsg(common_proto.LoginFromClient, &common_proto.LoginReq{
		AccountType: consts.TokenLogin,
		Account:     account,
		Game:        "elite-star",
	})
}

func (bot *ClientBot) sendGameMsg(cmd uint16, data []byte) {
	bot.SendMsg(common_proto.GameMethodFromClient, &common_proto.GameReq{
		Cmd:   cmd,
		Param: data,
	})
}
