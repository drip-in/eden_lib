package slices

// IndexOfString gets the index at which the first occurrence of an string value is found in array or return -1
// if the value cannot be found
func IndexOfString(arr []string, x string) int {
	for i := 0; i < len(arr); i++ {
		if arr[i] == x {
			return i
		}
	}
	return -1
}

// IndexOfInt gets the index at which the first occurrence of an int value is found in array or return -1
// if the value cannot be found
func IndexOfInt(arr []int, x int) int {
	for i := 0; i < len(arr); i++ {
		if arr[i] == x {
			return i
		}
	}
	return -1
}

// IndexOfInt8 gets the index at which the first occurrence of an int8 value is found in array or return -1
// if the value cannot be found
func IndexOfInt8(arr []int8, x int8) int {
	for i := 0; i < len(arr); i++ {
		if arr[i] == x {
			return i
		}
	}
	return -1
}

// IndexOfInt16 gets the index at which the first occurrence of an int16 value is found in array or return -1
// if the value cannot be found
func IndexOfInt16(arr []int16, x int16) int {
	for i := 0; i < len(arr); i++ {
		if arr[i] == x {
			return i
		}
	}
	return -1
}

// IndexOfInt32 gets the index at which the first occurrence of an int32 value is found in array or return -1
// if the value cannot be found
func IndexOfInt32(arr []int32, x int32) int {
	for i := 0; i < len(arr); i++ {
		if arr[i] == x {
			return i
		}
	}
	return -1
}

// IndexOfInt64 gets the index at which the first occurrence of an int64 value is found in array or return -1
// if the value cannot be found
func IndexOfInt64(arr []int64, x int64) int {
	for i := 0; i < len(arr); i++ {
		if arr[i] == x {
			return i
		}
	}
	return -1
}

// IndexOfUint gets the index at which the first occurrence of an uint value is found in array or return -1
// if the value cannot be found
func IndexOfUint(arr []uint, x uint) int {
	for i := 0; i < len(arr); i++ {
		if arr[i] == x {
			return i
		}
	}
	return -1
}

// IndexOfUint8 gets the index at which the first occurrence of an uint8 value is found in array or return -1
// if the value cannot be found
func IndexOfUint8(arr []uint8, x uint8) int {
	for i := 0; i < len(arr); i++ {
		if arr[i] == x {
			return i
		}
	}
	return -1
}

// IndexOfUint16 gets the index at which the first occurrence of an uint16 value is found in array or return -1
// if the value cannot be found
func IndexOfUint16(arr []uint16, x uint16) int {
	for i := 0; i < len(arr); i++ {
		if arr[i] == x {
			return i
		}
	}
	return -1
}

// IndexOfUint32 gets the index at which the first occurrence of an uint32 value is found in array or return -1
// if the value cannot be found
func IndexOfUint32(arr []uint32, x uint32) int {
	for i := 0; i < len(arr); i++ {
		if arr[i] == x {
			return i
		}
	}
	return -1
}

// IndexOfUint64 gets the index at which the first occurrence of an uint64 value is found in array or return -1
// if the value cannot be found
func IndexOfUint64(arr []uint64, x uint64) int {
	for i := 0; i < len(arr); i++ {
		if arr[i] == x {
			return i
		}
	}
	return -1
}

// IndexOfFloat32 gets the index at which the first occurrence of an float32 value is found in array or return -1
// if the value cannot be found
func IndexOfFloat32(arr []float32, x float32) int {
	for i := 0; i < len(arr); i++ {
		if arr[i] == x {
			return i
		}
	}
	return -1
}

// IndexOfFloat64 gets the index at which the first occurrence of an float64 value is found in array or return -1
// if the value cannot be found
func IndexOfFloat64(arr []float64, x float64) int {
	for i := 0; i < len(arr); i++ {
		if arr[i] == x {
			return i
		}
	}
	return -1
}
