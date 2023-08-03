package util

func IMax(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func IAbs(i int) int {
	if i < 0 {
		return -i
	}
	return i
}
