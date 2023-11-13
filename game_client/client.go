package game_client

import (
	context2 "context"
	"github.com/robfig/cron/v3"
	"go-one/common/common_proto"
	"go-one/common/consts"
	"go-one/common/json"
	"go-one/common/log"
	"go-one/common/network"
	"go-one/common/pktconn"
	"go-one/common/utils"
	"go-one/game/common"
	"go-one/game/entity"
	"net"
	"sync"

	"fmt"

	"github.com/xtaci/kcp-go"
	"golang.org/x/net/websocket"
)

var ClientContext = make(map[int64]*Client)

type Client struct {
	sync.Mutex
	ServerHost string

	ID int64

	conn        *pktconn.PacketConn
	packetQueue chan *pktconn.Packet
	crontab     *cron.Cron

	Position entity.Vector3
	Yaw      common.Yaw
	Speed    common.Speed

	I IClient
}

type IClient interface {
	OnCreated(client *Client)
	EnterGameParamWrapper(client *Client) *common_proto.EnterGameReq
	OnEnterGameSuccess(client *Client, resp *common_proto.EnterGameResp)
	OnJoinScene(client *Client, joinSceneResp *common_proto.JoinSceneResp)
}

func (c *Client) Init(ID int64) *Client {
	c.ID = ID
	c.packetQueue = make(chan *pktconn.Packet)
	c.crontab = cron.New(cron.WithSeconds())
	c.I.OnCreated(c)
	if Config.ServerConfig.UseLoadBalancer {
		param := make(map[string]string)
		param["partition"] = Config.ServerConfig.Partition
		param["entityID"] = utils.ToString(ID)

		resp, err := utils.Get(Config.ServerConfig.LoadBalancerUrl, param)
		if err != nil {
			panic(err)
		}
		var r result
		err = json.UnmarshalFromString(resp, &r)
		if err != nil {
			panic(err)
		}
		if Config.ServerConfig.Websocket {
			c.ServerHost = r.Data.WsAddr
		} else {
			c.ServerHost = r.Data.TcpAddr
		}
	} else {
		c.ServerHost = Config.ServerConfig.ServerHost
	}
	return c
}

type result struct {
	Code string         `json:"code"`
	Data chooseGateResp `json:"data"`
}
type chooseGateResp struct {
	WsAddr  string
	TcpAddr string
	Version string
}

func (c *Client) String() string {
	return fmt.Sprintf("Client<%d>", c.ID)
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

	c.crontab.AddFunc("@every 10s", func() {
		packet := pktconn.NewPacket()
		packet.WriteUint16(common_proto.HeartbeatFromClient)
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

	conn, err := network.ConnectTCP(c.ServerHost)
	if err == nil {
		conn.(*net.TCPConn).SetWriteBuffer(64 * 1024)
		conn.(*net.TCPConn).SetReadBuffer(64 * 1024)
	}
	return conn, err
}

func (c *Client) connectServerByKCP() (net.Conn, error) {

	conn, err := kcp.DialWithOptions(c.ServerHost, nil, 0, 0)
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

	origin := fmt.Sprintf("%s://%s/", originProto, c.ServerHost)
	wsaddr := fmt.Sprintf("%s://%s/ws", wsProto, c.ServerHost)

	return websocket.Dial(wsaddr, "", origin)
}

func (c *Client) recvLoop() {
	err := c.conn.ReceiveChan(c.packetQueue)
	log.Error(err)
}

func (c *Client) loop() {
	for {
		select {
		case pkt := <-c.packetQueue:
			c.handlePacket(pkt)
			pkt.Release()
			//break

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
	/*if msgType != 2001 {
		log.Infof("handlePacket: %d", msgType)
	}*/
	switch msgType {
	case common_proto.ConnectionSuccessFromServer:
		c.enterGame()
	case common_proto.EnterGameClientAck:
		loginResp := &common_proto.EnterGameResp{}
		packet.ReadData(loginResp)
		log.Infof("登录结果,EntityID:%d,game:%s", loginResp.EntityID, loginResp.Game)
		c.ID = loginResp.EntityID
		c.I.OnEnterGameSuccess(c, loginResp)

		ClientContext[c.ID] = c
	case common_proto.BroadcastFromServer:
		broadcastMsg := &common_proto.GateBroadcastMsg{}
		packet.ReadData(broadcastMsg)

		c.BroadcastMsgHandler(broadcastMsg)
	case common_proto.GameMethodFromClientAck:
		gameResp := &common_proto.GameResp{}
		packet.ReadData(gameResp)
		processor := ProcessorContext[gameResp.Cmd]
		if processor == nil {
			log.Warnf("未找到处理器:%d,resp: %s", gameResp.Cmd, string(gameResp.Data))
			return
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
	c.SendMsg(common_proto.EnterGameFromClient, c.I.EnterGameParamWrapper(c))
	log.Infof("发送登录消息:%s", c)
}

func (c *Client) BroadcastMsgHandler(msg *common_proto.GateBroadcastMsg) {
	log.Infof("收到广播消息:%s", msg.Data)
}

func (c *Client) SendGameData(cmd uint16, data any) {
	c.SendMsg(common_proto.GameMethodFromClient, &common_proto.GameReq{
		Cmd:   cmd,
		Param: PackMsg(data),
	})
}
