package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/ambitiousmice/go-one/common/json"
	"github.com/ambitiousmice/go-one/common/utils"
	"io/ioutil"
)

func main() {
	// 定义字符串数组
	nums := []int{56, 131, 176, 210, 176, 320, 563, 742, 421, 312}
	for i := 0; i < len(nums); i++ {
		strings := make([]string, nums[i])

		// 按照规则生成字符串
		for i := 0; i < len(strings); i++ {
			strings[i] = generateFormattedString()
		}

		// 将字符串数组转换成JSON
		jsonData, err := json.MarshalToString(strings)
		if err != nil {
			panic(err)
		}

		filePath := utils.ToString(i+1) + ".json"
		err = ioutil.WriteFile(filePath, []byte(jsonData), 0644)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("Output file was saved successfully in", filePath)
		}
	}

}

// 生成格式化随机字符串的函数
func generateFormattedString() string {
	part1 := generateRandomString(6)  // 6个字符的随机字符串
	part2 := generateRandomString(32) // 32个字符的随机字符串
	part3 := generateRandomString(4)  // 4个字符的随机字符串

	return part1 + "." + part2 + "." + part3
}

// 生成随机字符串的函数
func generateRandomString(length int) string {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(bytes)[:length]
}
