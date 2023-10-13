package dispatcher

import (
	"errors"
	"fmt"
	"github.com/robfig/cron/v3"
	"go-one/common/consts"
	"go-one/common/log"
	"go-one/common/pktconn"
	"sync"
	"sync/atomic"
)

type GameDispatcher struct {
	sync.Mutex
	game          string
	gameClusterID uint8
	gameHost      string
	gamePort      uint64
	gameOnlineNum int32
	status        int8
	pollingIndex  uint64

	cron     *cron.Cron
	channels map[uint8]*GameDispatcherChannel
}

func NewGameDispatcher(game string, gameClusterID uint8, gameHost string, gamePort uint64) *GameDispatcher {
	return &GameDispatcher{
		game:          game,
		gameClusterID: gameClusterID,
		gameHost:      gameHost,
		gamePort:      gamePort,
		status:        consts.DispatcherStatusInit,

		cron:     cron.New(cron.WithSeconds()),
		channels: make(map[uint8]*GameDispatcherChannel),
	}
}

func (gd *GameDispatcher) String() string {
	return fmt.Sprintf("GameDispatcher<%s><%d>", gd.game, gd.gameClusterID)
}

func (gd *GameDispatcher) Run() {
	for _, channel := range gd.channels {
		go channel.Run()
	}

	gd.cron.AddFunc("@every 5s", func() {
		gd.checkChannelHealth()
	})

	gd.cron.Start()

	log.Infof("GameDispatcher<%s><%d> started", gd.game, gd.gameClusterID)
}

func (gd *GameDispatcher) checkChannelHealth() {
	var healthCount = 0
	for _, channel := range gd.channels {
		if channel.getStatus() == consts.DispatcherChannelStatusHealth {
			healthCount++
			continue
		}

		channel.ReRun()
	}

	if healthCount == 0 {
		gd.status = consts.DispatcherStatusUnHealth
	}
}

func (gd *GameDispatcher) ForwardMsg(entityID int64, packet *pktconn.Packet) error {
	packet.WriteInt64(entityID)

	pollingIndex := uint8(atomic.AddUint64(&gd.pollingIndex, 1) % uint64(len(gd.channels)))
	channel := gd.channels[pollingIndex]

	if channel == nil {
		log.Errorf("no available channel,")
		return errors.New("no available channel")
	}

	if channel.status == consts.DispatcherChannelStatusHealth {
		channel.Send(packet)
	} else {
		// TODO: 重试优化
		return errors.New("no available channel")
	}
	return nil
}

func (gd *GameDispatcher) closeAll() {
	gd.cron.Stop()
	for key, channel := range gd.channels {
		channel.stop()
		delete(gd.channels, key)
	}
}

func (gd *GameDispatcher) CloseOne(channelID uint8) {
	channel := gd.channels[channelID]

	channel.cron.Stop()
	channel.updateStatus(consts.DispatcherChannelStatusStop)
	err := channel.Close()
	log.Warnf("close channel<%d> failed: %s", channelID, err.Error())
	delete(gd.channels, channelID)
}

func (gd *GameDispatcher) GetGameClusterID() uint8 {
	return gd.gameClusterID
}
