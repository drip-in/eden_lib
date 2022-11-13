package slices

// Contains returns true if an element is present in a slice/array.
func Contains(arr interface{}, elem interface{}) bool {
	switch realArr := arr.(type) {
	case []string:
		return ContainsString(realArr, elem.(string))

	case []int:
		return ContainsInt(realArr, elem.(int))

	case []int8:
		return ContainsInt8(realArr, elem.(int8))

	case []int16:
		return ContainsInt16(realArr, elem.(int16))

	case []int32:
		return ContainsInt32(realArr, elem.(int32))

	case []int64:
		return ContainsInt64(realArr, elem.(int64))

	case []uint:
		return ContainsUint(realArr, elem.(uint))

	case []uint8:
		return ContainsUint8(realArr, elem.(uint8))

	case []uint16:
		return ContainsUint16(realArr, elem.(uint16))

	case []uint32:
		return ContainsUint32(realArr, elem.(uint32))

	case []uint64:
		return ContainsUint64(realArr, elem.(uint64))

	case []float32:
		return ContainsFloat32(realArr, elem.(float32))

	case []float64:
		return ContainsFloat64(realArr, elem.(float64))

	default:
	}
	return containsIFace(arr, elem)
}

// Find iterates over a collection, returning the first element predicate returns truthy for.
// params:
//   arr must be slice or array,
//   cb must be a function: func(val $type) bool.
func Find(arr interface{}, cb interface{}) (interface{}, int) {
	switch realArr := arr.(type) {
	case []string:
		return FindString(realArr, cb.(func(v string) bool))

	case []int:
		return FindInt(realArr, cb.(func(v int) bool))

	case []int8:
		return FindInt8(realArr, cb.(func(v int8) bool))

	case []int16:
		return FindInt16(realArr, cb.(func(v int16) bool))

	case []int32:
		return FindInt32(realArr, cb.(func(v int32) bool))

	case []int64:
		return FindInt64(realArr, cb.(func(v int64) bool))

	case []uint:
		return FindUint(realArr, cb.(func(v uint) bool))

	case []uint8:
		return FindUint8(realArr, cb.(func(v uint8) bool))

	case []uint16:
		return FindUint16(realArr, cb.(func(v uint16) bool))

	case []uint32:
		return FindUint32(realArr, cb.(func(v uint32) bool))

	case []uint64:
		return FindUint64(realArr, cb.(func(v uint64) bool))

	case []float32:
		return FindFloat32(realArr, cb.(func(v float32) bool))

	case []float64:
		return FindFloat64(realArr, cb.(func(v float64) bool))

	default:
	}
	return findIFace(arr, cb)
}

// IndexOf gets the index at which the first occurrence value is found in collection or return -1
// if the value cannot be found.
func IndexOf(arr interface{}, x interface{}) int {
	switch realArr := arr.(type) {
	case []string:
		return IndexOfString(realArr, x.(string))

	case []int:
		return IndexOfInt(realArr, x.(int))

	case []int8:
		return IndexOfInt8(realArr, x.(int8))

	case []int16:
		return IndexOfInt16(realArr, x.(int16))

	case []int32:
		return IndexOfInt32(realArr, x.(int32))

	case []int64:
		return IndexOfInt64(realArr, x.(int64))

	case []uint:
		return IndexOfUint(realArr, x.(uint))

	case []uint8:
		return IndexOfUint8(realArr, x.(uint8))

	case []uint16:
		return IndexOfUint16(realArr, x.(uint16))

	case []uint32:
		return IndexOfUint32(realArr, x.(uint32))

	case []uint64:
		return IndexOfUint64(realArr, x.(uint64))

	case []float32:
		return IndexOfFloat32(realArr, x.(float32))

	case []float64:
		return IndexOfFloat64(realArr, x.(float64))

	default:
	}
	return indexOfIFace(arr, x)
}
