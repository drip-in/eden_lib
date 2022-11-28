package slices

// ContainsString returns true if an string is present in a iteratee.
func ContainsString(s []string, v string) bool {
	for _, vv := range s {
		if vv == v {
			return true
		}
	}
	return false
}

// ContainsInt returns true if an int is present in a iteratee.
func ContainsInt(s []int, v int) bool {
	for _, vv := range s {
		if vv == v {
			return true
		}
	}
	return false
}

// ContainsInt8 returns true if an int8 is present in a iteratee.
func ContainsInt8(s []int8, v int8) bool {
	for _, vv := range s {
		if vv == v {
			return true
		}
	}
	return false
}

// ContainsInt16 returns true if an int16 is present in a iteratee.
func ContainsInt16(s []int16, v int16) bool {
	for _, vv := range s {
		if vv == v {
			return true
		}
	}
	return false
}

// ContainsInt32 returns true if an int32 is present in a iteratee.
func ContainsInt32(s []int32, v int32) bool {
	for _, vv := range s {
		if vv == v {
			return true
		}
	}
	return false
}

// ContainsInt64 returns true if an int64 is present in a iteratee.
func ContainsInt64(s []int64, v int64) bool {
	for _, vv := range s {
		if vv == v {
			return true
		}
	}
	return false
}

// ContainsUint returns true if an uint is present in a iteratee.
func ContainsUint(s []uint, v uint) bool {
	for _, vv := range s {
		if vv == v {
			return true
		}
	}
	return false
}

// ContainsUint8 returns true if an uint8 is present in a iteratee.
func ContainsUint8(s []uint8, v uint8) bool {
	for _, vv := range s {
		if vv == v {
			return true
		}
	}
	return false
}

// ContainsUint16 returns true if an uint16 is present in a iteratee.
func ContainsUint16(s []uint16, v uint16) bool {
	for _, vv := range s {
		if vv == v {
			return true
		}
	}
	return false
}

// ContainsUint32 returns true if an uint32 is present in a iteratee.
func ContainsUint32(s []uint32, v uint32) bool {
	for _, vv := range s {
		if vv == v {
			return true
		}
	}
	return false
}

// ContainsUint64 returns true if an uint64 is present in a iteratee.
func ContainsUint64(s []uint64, v uint64) bool {
	for _, vv := range s {
		if vv == v {
			return true
		}
	}
	return false
}

// ContainsFloat32 returns true if an float32 is present in a iteratee.
func ContainsFloat32(s []float32, v float32) bool {
	for _, vv := range s {
		if vv == v {
			return true
		}
	}
	return false
}

// ContainsFloat64 returns true if an float64 is present in a iteratee.
func ContainsFloat64(s []float64, v float64) bool {
	for _, vv := range s {
		if vv == v {
			return true
		}
	}
	return false
}
