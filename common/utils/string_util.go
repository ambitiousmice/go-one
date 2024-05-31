package utils

import (
	"encoding/json"
	"fmt"
	"strconv"
	"unsafe"
)

func ToString(value interface{}) string {
	// interface 转 string
	var key string
	if value == nil {
		return key
	}

	switch value.(type) {
	case float64:
		ft := value.(float64)
		key = strconv.FormatFloat(ft, 'f', -1, 64)
	case float32:
		ft := value.(float32)
		key = strconv.FormatFloat(float64(ft), 'f', -1, 64)
	case int:
		it := value.(int)
		key = strconv.Itoa(it)
	case uint:
		it := value.(uint)
		key = strconv.Itoa(int(it))
	case int8:
		it := value.(int8)
		key = strconv.Itoa(int(it))
	case uint8:
		it := value.(uint8)
		key = strconv.Itoa(int(it))
	case int16:
		it := value.(int16)
		key = strconv.Itoa(int(it))
	case uint16:
		it := value.(uint16)
		key = strconv.Itoa(int(it))
	case int32:
		it := value.(int32)
		key = strconv.Itoa(int(it))
	case uint32:
		it := value.(uint32)
		key = strconv.Itoa(int(it))
	case int64:
		it := value.(int64)
		key = strconv.FormatInt(it, 10)
	case uint64:
		it := value.(uint64)
		key = strconv.FormatUint(it, 10)
	case string:
		key = value.(string)
	case []byte:
		key = string(value.([]byte))
	default:
		newValue, _ := json.Marshal(value)
		key = string(newValue)
	}

	return key
}

func ToInt(i interface{}) (num int, err error) {
	switch i.(type) {
	case string:
		num, err = strconv.Atoi(i.(string))
	}
	return
}

// ToInt32
/**
 * @Description: 将任意类型转为int32类型
 * @param i
 * @return num
 * @return err
 */
func ToInt32(i interface{}) (num int32, err error) {
	switch i.(type) {
	case int:
		num = int32(i.(int))
	case int32:
		num = i.(int32)
	case int64:
		// 有可能造成精度丢失
		num = int32(i.(int64))
	case float32:
		// 有可能造成精度丢失
		num = int32(i.(float32))
	case float64:
		// 有可能造成精度丢失
		num = int32(i.(float64))
	case string:
		n, e := strconv.Atoi(i.(string))
		num = int32(n)
		err = e
	default:
		panic("该类型暂不支持")
	}
	return
}

// ToInt64
/**
 * @Description: 将任意类型转为int64类型
 * @param i
 * @return num
 * @return err
 */
func ToInt64(i interface{}) (num int64, err error) {
	switch i.(type) {
	case int:
		num = int64(i.(int))
	case int32:
		num = int64(i.(int32))
	case int64:
		num = i.(int64)
	case float32:
		num = int64(i.(float32))
	case float64:
		num = int64(i.(float64))
	case string:
		num, err = strconv.ParseInt(i.(string), 10, 64)
	default:
		panic("该类型暂不支持")
	}
	return
}

// ToFloat32
/**
 * @Description: 将任意类型转为float32类型
 * @param i
 * @return num
 * @return err
 */
func ToFloat32(i interface{}) (num float32, err error) {
	switch i.(type) {
	case string:
		// string无法直接转换float32，只能先转换为float64，再通过float64转float32
		var num64 float64
		num64, err = strconv.ParseFloat(i.(string), 32)
		num = float32(num64)
	case int:
		num = float32(i.(int))
	case int32:
		num = float32(i.(int32))
	case int64:
		num = float32(i.(int64))
	case float32:
		num = i.(float32)
	case float64:
		// 可能造成精度丢失
		num = float32(i.(float64))
	default:
		panic("该类型暂不支持")
	}
	return
}

// ToFloat64
/**
 * @Description: 将任意类型转为float64类型
 * @param i
 * @return num
 * @return err
 */
func ToFloat64(i interface{}) (num float64, err error) {
	switch i.(type) {
	case string:
		num, err = strconv.ParseFloat(i.(string), 64)
	case int:
		num = float64(i.(int))
	case int32:
		num = float64(i.(int32))
	case int64:
		num = float64(i.(int64))
	case float32:
		num = float64(i.(float32))
	case float64:
		num = i.(float64)
	default:
		panic("该类型暂不支持")
	}
	return
}

func ToFloat64WithOutErr(i interface{}) (num float64) {
	switch i.(type) {
	case string:
		num, _ = strconv.ParseFloat(i.(string), 64)
	case int:
		num = float64(i.(int))
	case int32:
		num = float64(i.(int32))
	case int64:
		num = float64(i.(int64))
	case float32:
		num = float64(i.(float32))
	case float64:
		num = i.(float64)
	default:
		panic("该类型暂不支持")
	}
	return
}

// ToByteArray
/**
 *  @Description: 转换为[]byte
 *  @param i
 *  @return b
 */
func ToByteArray(i interface{}) (b []byte) {
	switch i.(type) {
	case string:
		str := i.(string)
		return *(*[]byte)(unsafe.Pointer(&str))
	default:
		panic("该类型暂不支持")
	}
	return
}

func ByteArrayToJsonStruct(b []byte, i interface{}) (interface{}, error) {
	err := json.Unmarshal(b, &i)
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
		return nil, err
	}
	return i, nil
}

func Contains[T comparable](slice []T, value T) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

func ContainsSlice[T comparable](slice []T, slice2 []T) bool {
	if len(slice) == 0 || len(slice2) == 0 {
		return false
	}
	for _, v := range slice2 {
		for _, v2 := range slice {
			if v == v2 {
				return true
			}
		}
	}
	return false
}
