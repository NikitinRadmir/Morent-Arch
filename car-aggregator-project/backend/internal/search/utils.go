package search

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// min3 returns the minimum of three integers
func min3(a, b, c int) int {
	return min(min(a, b), c)
}

// abs returns the absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}