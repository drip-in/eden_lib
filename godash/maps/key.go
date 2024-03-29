package maps

import "reflect"

// Keys return map's keys as unordered slice.
func Keys(mp interface{}) interface{} {
	mpValue := reflect.ValueOf(mp)
	mpLen := mpValue.Len()
	sliceType := reflect.SliceOf(mpValue.Type().Key())
	if mpLen == 0 {
		return reflect.Zero(sliceType).Interface()
	}
	keys := reflect.MakeSlice(sliceType, 0, mpValue.Len())
	iter := mpValue.MapRange()
	for iter.Next() {
		keys = reflect.Append(keys, iter.Key())
	}
	return keys.Interface()
}
