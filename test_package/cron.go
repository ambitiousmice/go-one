package main

import (
	"encoding/json"
	"fmt"
	"go-one/test_package/t"
	"log"

	"github.com/golang/protobuf/proto"
)

func main() {
	// 创建包含100个AOISyncInfo对象的切片
	aoiSlice := make([]*t.AOISyncInfo, 100)
	for i := 0; i < 100; i++ {
		aoi := &t.AOISyncInfo{
			EntityId: int64(i),
			X:        float32(i),
			Y:        float32(i),
			Z:        float32(i),
			Yaw:      float32(i),
			Speed:    float32(i),
		}
		aoiSlice[i] = aoi
	}

	// 进行JSON序列化
	jsonData, err := json.Marshal(aoiSlice)
	if err != nil {
		log.Fatal(err)
	}
	jsonSize := len(jsonData)
	println(string(jsonData))
	// 进行Protocol Buffers序列化
	protoData, err := proto.Marshal(&t.AOISyncInfoList{AoiSyncInfo: aoiSlice})
	if err != nil {
		log.Fatal(err)
	}
	protoSize := len(protoData)

	// 计算大小差异倍数
	ratio := float64(protoSize) / float64(jsonSize)

	fmt.Printf("JSON序列化大小: %d bytes\n", jsonSize)
	fmt.Printf("Protocol Buffers序列化大小: %d bytes\n", protoSize)
	fmt.Printf("大小差异倍数: %.2f倍\n", ratio)
}
