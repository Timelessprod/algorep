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
