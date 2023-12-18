package main

import (
	"go-one/common/log"
	"go-one/demo/im/proto"
	"go-one/game_client"
	"time"
)

type SubscribeRoomProcessor struct {
}

func (p *SubscribeRoomProcessor) Process(client *game_client.Client, param []byte) {
	subscribeRoomResp := &proto.SubscribeRoomResp{}
	game_client.UnPackMsg(param, subscribeRoomResp)

	log.Infof(" %s subscribe room success: %d", client, subscribeRoomResp.RoomID)

	/*go func() {
		pushMessageReq := &proto.PushMessageReq{
			RoomID: 1001,
			//Data:    "{\"test\":1}",
			Msg: "{\"test\":1}",
		}
		for {
			currentTime := time.Now()

			// 将时间转换为字符串，使用指定的格式
			formattedTime := currentTime.Format("2006-01-02 15:04:05")

			// 打印格式化后的时间字符串
			pushMessageReq.Msg = formattedTime
			for i := 0; i < 10; i++ {
				client.SendGameData(proto.PushRoomMessage, pushMessageReq)
			}
			time.Sleep(1 * time.Second)
		}
	}()*/

	go func() {
		for {
			currentTime := time.Now()

			// 将时间转换为字符串，使用指定的格式
			formattedTime := currentTime.Format("2006-01-02 15:04:05")
			for i := 0; i < 40; i++ {
				client.SendGameData(proto.PushRoomMessage, &proto.PushMessageReq{
					RoomID: 1001,
					Msg:    formattedTime,
				})
			}
			time.Sleep(1 * time.Second)
		}
	}()

	/*go func() {
		for {
			client.SendGameData(proto.PushOneMessage, &proto.PushMessageReq{
				To:  5,
				Data: "hello world:" + utils.ToString(client.ID),
			})
			randomInt := rand.Intn(3)
			sleepTime := time.Duration(randomInt) * time.Second
			time.Sleep(sleepTime)
		}
	}()*/

}

func (p *SubscribeRoomProcessor) GetCmd() uint16 {
	return proto.SubscribeRoomAck
}

type MessageAckProcessor struct {
}

func (p *MessageAckProcessor) Process(client *game_client.Client, param []byte) {
	log.Infof("receive message: %s", string(param))
}

func (p *MessageAckProcessor) GetCmd() uint16 {
	return proto.MessageAck
}
