package utils

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
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

// Post 处理post请求
func PostWithHeader(url string, param interface{}, headers map[string]string, respData interface{}) error {
	data, err := json.Marshal(param)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	req.Header.Add("Content-Type", "application/json")
	if headers != nil {
		for k, v := range headers {
			req.Header.Add(k, v)
		}
	}

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

func Get(requestUrl string, queryParams map[string]string) (string, error) {
	// 创建查询参数字符串
	query := url.Values{}
	for key, value := range queryParams {
		query.Add(key, value)
	}

	// 将查询参数附加到URL
	fullURL := requestUrl + "?" + query.Encode()

	// 发起带查询参数的GET请求
	resp, err := http.Get(fullURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// 返回响应内容
	return string(body), nil
}

func GetWithHeader(requestUrl string, queryParams map[string]string, headers map[string]string, respData interface{}) error {
	// 创建查询参数字符串
	query := url.Values{}
	for key, value := range queryParams {
		query.Add(key, value)
	}

	// 将查询参数附加到URL
	fullURL := requestUrl + "?" + query.Encode()

	req, err := http.NewRequest("GET", fullURL, nil)
	if headers != nil {
		for k, v := range headers {
			req.Header.Add(k, v)
		}
	}

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
