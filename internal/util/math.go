package util

import "math"

func DivideAndFloorI32(a int32, b int32) int32 {
	return int32(math.Floor(float64(a) / float64(b)))
}

func I32Abs(i int32) int32 {
	if i < 0 {
		return -i
	} else {
		return i
	}
}
