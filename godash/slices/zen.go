package slices

// The reason we call the filename as "zen":
// it's the utility for slice(with as interface) operation based on reflect,
// Using reflection techniques to complete the work is like a guru who master zen.

import (
	"bytes"
	"fmt"
	"math/rand"
	"reflect"
)

var (
	emptyStruct = struct{}{}
)

// region --- Elem
func containsIFace(arr, elem interface{}) bool {
	inValue := reflect.ValueOf(arr)
	inType := inValue.Type()
	inKind := inType.Kind()
	if inKind != reflect.Slice && inKind != reflect.Array {
		panic(
			fmt.Sprintf(
				"Type %s is not supported by Contains, supported types are Slice, Array",
				inType.String(),
			),
		)
	}
	for i := 0; i < inValue.Len(); i++ {
		expected := inValue.Index(i).Interface()
		if reflect.DeepEqual(expected, elem) {
			return true
		}
	}
	return false
}

// Find iterates over elements of collection, returning the first
// element predicate returns truthy for.
func findIFace(arr, predicate interface{}) (iVal interface{}, index int) {
	_, arrValue, funcValue := parseCollectionAndPredicate(arr, predicate)
	for index = 0; index < arrValue.Len(); index++ {
		elem := arrValue.Index(index)
		result := funcValue.Call([]reflect.Value{elem})[0].Interface().(bool)
		if result {
			return elem.Interface(), index
		}
	}
	return nil, -1
}

func indexOfIFace(arr, x interface{}) int {
	arrValue := reflect.ValueOf(arr)
	for i := 0; i < arrValue.Len(); i++ {
		elem := arrValue.Index(i).Interface()
		if reflect.DeepEqual(elem, x) {
			return i
		}
	}
	return -1
}

func lastIndexOfIFace(arr, x interface{}) int {
	arrValue := reflect.ValueOf(arr)
	for i := arrValue.Len() - 1; i >= 0; i-- {
		elem := arrValue.Index(i).Interface()
		if reflect.DeepEqual(elem, x) {
			return i
		}
	}
	return -1
}

// endregion

// region --- Slice

func distinctIFace(arr interface{}) interface{} {
	arrValue := reflect.ValueOf(arr)
	length := arrValue.Len()
	seen := make(map[interface{}]struct{})
	// We prefer lowest time cost first
	resultSlice := reflect.MakeSlice(arrValue.Type(), 0, prepareDistinctLength(length))
	for i := 0; i < length; i++ {
		elemValue := arrValue.Index(i)
		v := elemValue.Interface()
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = emptyStruct
		resultSlice = reflect.Append(resultSlice, elemValue)
	}
	return resultSlice.Interface()
}

func distinctIFaceInplace(arr interface{}) interface{} {
	arrValue := reflect.ValueOf(arr)
	length := arrValue.Len()
	seen := make(map[interface{}]bool)
	j := 0
	for i := 0; i < length; i++ {
		elemVal := arrValue.Index(i)
		v := elemVal.Interface()
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = true
		arrValue.Index(j).Set(elemVal)
		j++
	}
	return arrValue.Slice(0, j).Interface()
}

func distinctByIFace(arr, selector interface{}) interface{} {
	selectorValue := parseDistinctSelector(selector)
	return filterIFace(arr, buildFilterFuncForDistinctBy(selectorValue))
}

func distinctByIFaceInplace(arr, selector interface{}) interface{} {
	selectorValue := parseDistinctSelector(selector)
	return filterIFaceInplace(arr, buildFilterFuncForDistinctBy(selectorValue))
}

func filterIFace(arr, predicate interface{}) interface{} {
	arrType, arrValue, funcValue := parseCollectionAndPredicate(arr, predicate)
	resultSliceType := reflect.SliceOf(arrType.Elem())
	// MakeSlice takes a slice kind type, and makes a slice.
	resultSlice := reflect.MakeSlice(resultSliceType, 0, 0)
	for i := 0; i < arrValue.Len(); i++ {
		elem := arrValue.Index(i)
		result := funcValue.Call([]reflect.Value{elem})[0].Interface().(bool)
		if result {
			resultSlice = reflect.Append(resultSlice, elem)
		}
	}
	return resultSlice.Interface()
}

func filterIFaceInplace(arr, predicate interface{}) interface{} {
	_, arrValue, funcValue := parseCollectionAndPredicate(arr, predicate)
	index, cursor, length := 0, -1, arrValue.Len()
	for index < length {
		elem := arrValue.Index(index)
		result := funcValue.Call([]reflect.Value{elem})[0].Interface().(bool)
		if !result {
			index++
			continue
		}
		if cursor+1 == index {
			index++
			cursor++
			continue
		}
		// move val forward
		arrValue.Index(cursor + 1).Set(elem)
		cursor++
		index++
	}
	return arrValue.Slice(0, cursor+1).Interface()
}

func removeIFace(arr, x interface{}) interface{} {
	arrType := reflect.TypeOf(arr)
	arrValue := reflect.ValueOf(arr)
	resultSliceType := reflect.SliceOf(arrType.Elem())
	// MakeSlice takes a slice kind type, and makes a slice.
	resultSlice := reflect.MakeSlice(resultSliceType, 0, 0)
	for i := 0; i < arrValue.Len(); i++ {
		elem := arrValue.Index(i)
		if !reflect.DeepEqual(elem.Interface(), x) {
			resultSlice = reflect.Append(resultSlice, elem)
		}
	}
	return resultSlice.Interface()
}

func removeIFaceInplace(arr, x interface{}) interface{} {
	arrValue := reflect.ValueOf(arr)
	index, cursor, length := 0, -1, arrValue.Len()
	for index < length {
		elem := arrValue.Index(index)
		if reflect.DeepEqual(elem.Interface(), x) {
			index++
			continue
		}
		if cursor+1 == index {
			index++
			cursor++
			continue
		}
		// move val forward
		arrValue.Index(cursor + 1).Set(elem)
		cursor++
		index++
	}
	return arrValue.Slice(0, cursor+1).Interface()
}

func removeAllIFace(arr, removeArr interface{}) interface{} {
	removeArrValue := reflect.ValueOf(removeArr)
	removeLength := removeArrValue.Len()
	seen := make(map[interface{}]struct{}, removeLength)
	for i := 0; i < removeLength; i++ {
		seen[removeArrValue.Index(i).Interface()] = emptyStruct
	}

	arrValue := reflect.ValueOf(arr)
	resultSlice := reflect.MakeSlice(arrValue.Type(), 0, 0)
	length := arrValue.Len()
	for i := 0; i < length; i++ {
		elemValue := arrValue.Index(i)
		v := elemValue.Interface()
		if _, exists := seen[v]; !exists {
			resultSlice = reflect.Append(resultSlice, elemValue)
		}
	}
	return resultSlice.Interface()
}

func removeAllIFaceInplace(arr, removeArr interface{}) interface{} {
	removeArrValue := reflect.ValueOf(removeArr)
	removeLength := removeArrValue.Len()
	seen := make(map[interface{}]struct{}, removeLength)
	for i := 0; i < removeLength; i++ {
		seen[removeArrValue.Index(i).Interface()] = emptyStruct
	}

	var exists bool
	arrValue := reflect.ValueOf(arr)
	index, cursor, length := 0, -1, arrValue.Len()
	for index < length {
		elem := arrValue.Index(index)
		if _, exists = seen[elem.Interface()]; exists {
			index++
			continue
		}
		if cursor+1 == index {
			index++
			cursor++
			continue
		}
		// move val forward
		arrValue.Index(cursor + 1).Set(elem)
		cursor++
		index++
	}
	return arrValue.Slice(0, cursor+1).Interface()
}

func reverseIFace(arr interface{}) interface{} {
	arrType := reflect.TypeOf(arr)
	arrValue := reflect.ValueOf(arr)
	resultSliceType := reflect.SliceOf(arrType.Elem())

	length := arrValue.Len()
	// MakeSlice takes a slice kind type, and makes a slice.
	resultSlice := reflect.MakeSlice(resultSliceType, length, length)

	i, j := 0, length-1
	for ; i < length/2; i, j = i+1, j-1 {
		resultSlice.Index(i).Set(arrValue.Index(j))
		resultSlice.Index(j).Set(arrValue.Index(i))
	}
	if i == j {
		// len(s) mod 2 == 0
		resultSlice.Index(i).Set(arrValue.Index(i))
	}
	return resultSlice.Interface()
}

func reverseIFaceInplace(arr interface{}) interface{} {
	length := reflect.ValueOf(arr).Len()
	swap := reflect.Swapper(arr)
	for i, j := 0, length-1; i < length/2; i, j = i+1, j-1 {
		swap(i, j)
	}
	return arr
}

func shuffleIFace(arr interface{}) interface{} {
	arrValue := reflect.ValueOf(arr)
	length := arrValue.Len()
	resultSlice := reflect.MakeSlice(arrValue.Type(), length, length)
	reflect.Copy(resultSlice, arrValue)
	swap := reflect.Swapper(resultSlice.Interface())
	for i := 0; i < length; i++ {
		j := rand.Intn(i + 1)
		swap(i, j)
	}
	return resultSlice.Interface()
}

func shuffleIFaceInplace(arr interface{}) interface{} {
	arrValue := reflect.ValueOf(arr)
	length := arrValue.Len()
	swap := reflect.Swapper(arr)
	for i := 0; i < length; i++ {
		j := rand.Intn(i + 1)
		swap(i, j)
	}
	return arr
}

func shuffleIFaceByRand(arr interface{}, rr *rand.Rand) interface{} {
	arrValue := reflect.ValueOf(arr)
	length := arrValue.Len()
	resultSlice := reflect.MakeSlice(arrValue.Type(), length, length)
	reflect.Copy(resultSlice, arrValue)
	swap := reflect.Swapper(resultSlice.Interface())
	for i := 0; i < length; i++ {
		j := rr.Intn(i + 1)
		swap(i, j)
	}
	return resultSlice.Interface()
}

func shuffleIFaceByRandInplace(arr interface{}, rr *rand.Rand) interface{} {
	arrValue := reflect.ValueOf(arr)
	length := arrValue.Len()
	swap := reflect.Swapper(arr)
	for i := 0; i < length; i++ {
		j := rr.Intn(i + 1)
		swap(i, j)
	}
	return arr
}

// endregion

// region -- Utility

func getTypeAndKind(arr interface{}) (reflect.Type, reflect.Kind) {
	arrType := reflect.TypeOf(arr)
	kind := arrType.Kind()
	return arrType, kind
}

func isKindCollection(kind reflect.Kind) bool {
	return kind == reflect.Slice || kind == reflect.Array
}

func parseCollectionAndPredicate(arr, predicate interface{}) (
	arrType reflect.Type, arrValue, funcValue reflect.Value,
) {
	var kind reflect.Kind
	arrType, kind = getTypeAndKind(arr)
	if !isKindCollection(kind) {
		panic("arr must be slice or array")
	}
	funcValue = reflect.ValueOf(predicate)
	funcType := funcValue.Type()
	if !validateFuncType(funcType, 1, 1) {
		panic("predicate must be function")
	}
	if funcType.Out(0).Kind() != reflect.Bool {
		panic("predicate return argument should be a boolean")
	}
	arrValue = reflect.ValueOf(arr)
	return arrType, arrValue, funcValue
}

func buildFilterFuncForDistinctBy(selectorValue reflect.Value) func(elem interface{}) bool {
	seen := make(map[interface{}]struct{})
	return func(elem interface{}) bool {
		value := selectorValue.Call([]reflect.Value{reflect.ValueOf(elem)})[0]
		vIFace := value.Interface()
		if _, exists := seen[vIFace]; exists {
			return false
		}
		seen[vIFace] = emptyStruct
		return true
	}
}

func prepareDistinctLength(length int) int {
	if length <= 10240 {
		return length
	}
	if length <= 20480 {
		return 10240
	}
	return length / 2
}

func parseDistinctSelector(selector interface{}) (selectorValue reflect.Value) {
	selectorValue = reflect.ValueOf(selector)
	selectorType := selectorValue.Type()
	if selectorType.Kind() != reflect.Func {
		panic("selector must be function")
	}
	if selectorType.NumIn() != 1 {
		panic("selector params not match 1")
	}
	if selectorType.NumOut() != 1 {
		panic("selector return not match 1")
	}
	return selectorValue
}

func validateFuncType(funcType reflect.Type, numIn, outNum int) bool {
	if funcType.Kind() != reflect.Func {
		return false
	}
	if funcType.NumIn() != numIn {
		return false
	}
	if funcType.NumOut() != outNum {
		return false
	}
	return true
}

func objectsAreEqual(obj1, obj2 interface{}) bool {
	if obj1 == nil || obj2 == nil {
		return obj1 == obj2
	}

	exp, ok := obj1.([]byte)
	if !ok {
		return reflect.DeepEqual(obj1, obj2)
	}

	act, ok := obj2.([]byte)
	if !ok {
		return false
	}
	if exp == nil || act == nil {
		return exp == nil && act == nil
	}
	return bytes.Equal(exp, act)
}

func elementsMatchIFace(s1, s2 interface{}) bool {
	aValue := reflect.ValueOf(s1)
	bValue := reflect.ValueOf(s2)

	aLen := aValue.Len()
	bLen := bValue.Len()

	// Mark indexes in bValue that we already used
	visited := make([]bool, bLen)
	for i := 0; i < aLen; i++ {
		element := aValue.Index(i).Interface()
		found := false
		for j := 0; j < bLen; j++ {
			if visited[j] {
				continue
			}
			if objectsAreEqual(bValue.Index(j).Interface(), element) {
				visited[j] = true
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	for j := 0; j < bLen; j++ {
		if !visited[j] {
			return false
		}
	}

	return true
}

// endregion
