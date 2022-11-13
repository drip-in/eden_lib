package maps

import "reflect"

// Values return map's values as unordered slice.
func Values(mp interface{}) interface{} {
	mpValue := reflect.ValueOf(mp)
	mpLen := mpValue.Len()
	sliceType := reflect.SliceOf(mpValue.Type().Elem())
	if mpLen == 0 {
		return reflect.Zero(sliceType).Interface()
	}
	elems := reflect.MakeSlice(sliceType, 0, mpValue.Len())
	iter := mpValue.MapRange()
	for iter.Next() {
		elems = reflect.Append(elems, iter.Value())
	}
	return elems.Interface()
}
