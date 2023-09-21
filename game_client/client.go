package game_client

import (
	context2 "context"
	"github.com/robfig/cron/v3"
	"go-one/common/consts"
	"go-one/common/log"
	"go-one/common/network"
	"go-one/common/pktconn"
	"go-one/common/proto"
	"net"
	"sync"

	"fmt"

	"time"

	"github.com/xtaci/kcp-go"
	"golang.org/x/net/websocket"
)

var ClientContext = make(map[int64]*Client)

type Client struct {
	sync.Mutex

	id int64

	conn        *pktconn.PacketConn
	packetQueue chan *pktconn.Packet
	crontab     *cron.Cron

	I IClient
}

type IClient interface {
	OnCreated(client *Client)
	EnterGameParamWrapper(client *Client) *proto.EnterGameReq
	OnEnterGameSuccess(client *Client, resp *proto.EnterGameResp)
	OnJoinScene(client *Client, joinSceneResp *proto.JoinSceneResp)
}

func NewClient(i IClient) *Client {

	return &Client{
		packetQueue: make(chan *pktconn.Packet),
		crontab:     cron.New(cron.WithSeconds()),
		I:           i,
	}
}

func (c *Client) String() string {
	return fmt.Sprintf("Client<%d>", c.id)
}

func (c *Client) Run() {

	var netConn net.Conn
	netConn, err := c.connectServer()
	if err != nil {
		panic("connect server failed: " + err.Error())
	}

	log.Infof("%s connected: %s", c, netConn.RemoteAddr())

	netConn = pktconn.NewBufferedConn(netConn, consts.BufferedReadBufferSize, consts.BufferedWriteBufferSize)
	c.conn = pktconn.NewPacketConn(context2.Background(), netConn, c)
	defer c.conn.Close()
	c.crontab.Start()

	c.crontab.AddFunc("@every 2s", func() {
		packet := pktconn.NewPacket()
		packet.WriteUint16(proto.HeartbeatFromClient)
		c.conn.SendAndRelease(packet)
		//log.Infof("==============发送心跳包,packetQueue长度:%d", len(c.packetQueue))
	})
	// send handshake packet

	go c.recvLoop()

	c.loop()
}

func (c *Client) connectServer() (net.Conn, error) {
	if Config.ServerConfig.Websocket {
		return c.connectServerByWebsocket()
	} else if Config.ServerConfig.Kcp {
		return c.connectServerByKCP()
	}

	conn, err := network.ConnectTCP(net.JoinHostPort(Config.ServerConfig.IP, Config.ServerConfig.Port))
	if err == nil {
		conn.(*net.TCPConn).SetWriteBuffer(64 * 1024)
		conn.(*net.TCPConn).SetReadBuffer(64 * 1024)
	}
	return conn, err
}

func (c *Client) connectServerByKCP() (net.Conn, error) {

	serverAddr := net.JoinHostPort(Config.ServerConfig.IP, Config.ServerConfig.Port)
	conn, err := kcp.DialWithOptions(serverAddr, nil, 0, 0)
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

func (c *Client) connectServerByWebsocket() (net.Conn, error) {
	originProto := "http"
	wsProto := "ws"

	origin := fmt.Sprintf("%s://%s:%s/", originProto, Config.ServerConfig.IP, Config.ServerConfig.Port)
	wsaddr := fmt.Sprintf("%s://%s:%s/ws", wsProto, Config.ServerConfig.IP, Config.ServerConfig.Port)

	return websocket.Dial(wsaddr, "", origin)
}

func (c *Client) recvLoop() {
	err := c.conn.ReceiveChan(c.packetQueue)
	log.Error(err)
}

func (c *Client) loop() {
	ticker := time.Tick(time.Millisecond * 100)
	for {
		select {
		case pkt := <-c.packetQueue:
			c.handlePacket(pkt)
			pkt.Release()
			//break
		case <-ticker:

			break
		}
	}
}

func (c *Client) handlePacket(packet *pktconn.Packet) {
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
	case proto.ConnectionSuccessFromServer:
		c.enterGame()
		log.Infof("发送登录消息")
	case proto.EnterGameClientAck:
		loginResp := &proto.EnterGameResp{}
		packet.ReadData(loginResp)
		log.Infof("登录结果,EntityID:%d,game:%s", loginResp.EntityID, loginResp.Game)
		c.id = loginResp.EntityID
		c.I.OnEnterGameSuccess(c, loginResp)

		ClientContext[c.id] = c

	case proto.GameMethodFromClientAck:
		gameResp := &proto.GameResp{}
		packet.ReadData(gameResp)
		processor := ProcessorContext[gameResp.Cmd]
		if processor == nil {
			log.Warnf("未找到处理器:%d,resp: %s", gameResp.Cmd, string(gameResp.Data))
		}
		processor.Process(c, gameResp.Data)
	}

}

func (c *Client) SendMsg(msgType uint16, msg interface{}) {
	packet := pktconn.NewPacket()
	packet.WriteUint16(msgType)
	if msg != nil {
		packet.AppendData(msg)
	}
	c.conn.Send(packet)
	packet.Release()
}

func (c *Client) enterGame() {
	c.SendMsg(proto.EnterGameFromClient, c.I.EnterGameParamWrapper(c))
}

func (c *Client) sendGameMsg(cmd uint16, data []byte) {
	c.SendMsg(proto.GameMethodFromClient, &proto.GameReq{
		Cmd:   cmd,
		Param: data,
	})
}
