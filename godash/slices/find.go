package slices

// FindString iterates over a collection of string, returning the first
// string element predicate returns truthy for.
func FindString(s []string, cb func(v string) bool) (string, int) {
	for index, value := range s {
		result := cb(value)
		if result {
			return value, index
		}
	}
	return "", -1
}

// FindInt iterates over a collection of int, returning the first
// int element predicate returns truthy for.
func FindInt(s []int, cb func(v int) bool) (int, int) {
	for index, value := range s {
		result := cb(value)
		if result {
			return value, index
		}
	}
	return 0, -1
}

// FindInt8 iterates over a collection of int8, returning the first
// int8 element predicate returns truthy for.
func FindInt8(s []int8, cb func(v int8) bool) (int8, int) {
	for index, value := range s {
		result := cb(value)
		if result {
			return value, index
		}
	}
	return 0, -1
}

// FindInt16 iterates over a collection of int16, returning the first
// int16 element predicate returns truthy for.
func FindInt16(s []int16, cb func(v int16) bool) (int16, int) {
	for index, value := range s {
		result := cb(value)
		if result {
			return value, index
		}
	}
	return 0, -1
}

// FindInt32 iterates over a collection of int32, returning the first
// int32 element predicate returns truthy for.
func FindInt32(s []int32, cb func(v int32) bool) (int32, int) {
	for index, value := range s {
		result := cb(value)
		if result {
			return value, index
		}
	}
	return 0, -1
}

// FindInt64 iterates over a collection of int64, returning the first
// int64 element predicate returns truthy for.
func FindInt64(s []int64, cb func(v int64) bool) (int64, int) {
	for index, value := range s {
		result := cb(value)
		if result {
			return value, index
		}
	}
	return 0, -1
}

// FindUint iterates over a collection of uint, returning the first
// uint element predicate returns truthy for.
func FindUint(s []uint, cb func(v uint) bool) (uint, int) {
	for index, value := range s {
		result := cb(value)
		if result {
			return value, index
		}
	}
	return 0, -1
}

// FindUint8 iterates over a collection of uint8, returning the first
// uint8 element predicate returns truthy for.
func FindUint8(s []uint8, cb func(v uint8) bool) (uint8, int) {
	for index, value := range s {
		result := cb(value)
		if result {
			return value, index
		}
	}
	return 0, -1
}

// FindUint16 iterates over a collection of uint16, returning the first
// uint16 element predicate returns truthy for.
func FindUint16(s []uint16, cb func(v uint16) bool) (uint16, int) {
	for index, value := range s {
		result := cb(value)
		if result {
			return value, index
		}
	}
	return 0, -1
}

// FindUint32 iterates over a collection of uint32, returning the first
// uint32 element predicate returns truthy for.
func FindUint32(s []uint32, cb func(v uint32) bool) (uint32, int) {
	for index, value := range s {
		result := cb(value)
		if result {
			return value, index
		}
	}
	return 0, -1
}

// FindUint64 iterates over a collection of uint64, returning the first
// uint64 element predicate returns truthy for.
func FindUint64(s []uint64, cb func(v uint64) bool) (uint64, int) {
	for index, value := range s {
		result := cb(value)
		if result {
			return value, index
		}
	}
	return 0, -1
}

// FindFloat32 iterates over a collection of float32, returning the first
// float32 element predicate returns truthy for.
func FindFloat32(s []float32, cb func(v float32) bool) (float32, int) {
	for index, value := range s {
		result := cb(value)
		if result {
			return value, index
		}
	}
	return 0.0, -1
}

// FindFloat64 iterates over a collection of float64, returning the first
// float64 element predicate returns truthy for.
func FindFloat64(s []float64, cb func(v float64) bool) (float64, int) {
	for index, value := range s {
		result := cb(value)
		if result {
			return value, index
		}
	}
	return 0.0, -1
}
