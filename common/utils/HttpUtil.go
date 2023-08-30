package utils

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// Post 处理post请求
func Post(url string, param interface{}, respData interface{}) error {
	data, err := json.Marshal(param)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{} // 处理返回结果
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)  // 读取请求结果
	err = json.Unmarshal(body, &respData) // 将string 格式转成json格式

	if err != nil {
		return err
	}
	return nil
}
