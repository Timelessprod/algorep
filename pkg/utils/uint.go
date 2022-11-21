package utils

/***********
 ** Utils **
 ***********/

// Minimum of two uint32
func MinUint32(a uint32, b uint32) uint32 {
	if a <= b {
		return a
	}
	return b
}

// Maximum of two uint32
func MaxUint32(a uint32, b uint32) uint32 {
	if a >= b {
		return a
	}
	return b
}

// Get the index with the lowest value in a slice of uint32
func IndexMinUint32(values []uint32) uint32 {
	var minIndex uint32 = 0
	for i, v := range values {
		if v < values[minIndex] {
			minIndex = uint32(i)
		}
	}
	return minIndex
}
