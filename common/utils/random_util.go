package utils

import "math/rand"

func RandomInt64(min, max int64) int64 {
	if (max - min) <= 0 {
		return min
	}
	return rand.Int63n(max-min) + min
}

func RandomInt32(min, max int32) int32 {
	if (max - min) <= 0 {
		return min
	}
	return rand.Int31n(max-min) + min
}

func RandomInt(min, max int) int {
	if (max - min) <= 0 {
		return min
	}
	return rand.Intn(max-min) + min
}
