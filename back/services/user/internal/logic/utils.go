package logic

// max 返回两个整数中的最大值
func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

// clamp 将整数限制在指定范围内
func clamp(a, min, max int64) int64 {
	if a < min {
		return min
	}
	if a > max {
		return max
	}
	return a
}
