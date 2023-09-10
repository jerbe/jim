package utils

import (
	"github.com/jerbe/jim/errors"
	"reflect"
)

/**
  @author : Jerbe - The porter from Earth
  @time : 2023/8/27 16:12
  @describe :
*/

// IsNil 检测是否是真nil值
func IsNil(v any) bool {
	if v != nil {
		// 如果目标不是xx类型,则返回
		typ := reflect.TypeOf(v)
		if typ.Kind() != reflect.Pointer {
			return false
		}

		value := reflect.ValueOf(v)
		if !value.IsNil() {
			return false
		}
	}
	return true
}

// In 判断obj是否与target或者targets内的某个元素相等
// 如果 obj 等于 target 或者等于 targets 其中一项,则返回true
// 如果没匹配到其中一项,则返回false
func In(obj any, target any, targets ...any) bool {
	if obj == nil {
		if IsNil(target) {
			return true
		}

		for i := 0; i < len(targets); i++ {
			if IsNil(targets[i]) {
				return true
			}
		}
		return false
	}

	if obj == target {
		return true
	}

	for i := 0; i < len(targets); i++ {
		if obj == targets[i] {
			return true
		}
	}
	return false
}

// Equal 判断target跟targets内的所有项与val是否相等,如果全部target都与val相等
// 如果obj与target跟targets里面的所有数据都相等,则返回true,如果有其中一项不相等,则返回false
func Equal(obj any, target any, targets ...any) bool {
	if obj == nil {
		checkNil := func(v any) bool {
			if v != nil {
				// 如果目标不是xx类型,则返回
				typ := reflect.TypeOf(v)
				if typ.Kind() != reflect.Pointer {
					return false
				}

				value := reflect.ValueOf(v)
				if !value.IsNil() {
					return false
				}
			}
			return true
		}

		if !checkNil(target) {
			return false
		}

		for i := 0; i < len(targets); i++ {
			if !checkNil(targets[i]) {
				return false
			}
		}
		return true
	}

	if obj != target {
		return false
	}
	for i := 0; i < len(targets); i++ {
		if obj != targets[i] {
			return false
		}
	}
	return true
}

// SliceUnique 将切片进行唯一归类
func SliceUnique(data any, dst any) error {
	dataType := reflect.TypeOf(data)
	if dataType.Kind() == reflect.Ptr {
		dataType = dataType.Elem()
	}
	if dataType.Kind() != reflect.Slice {
		return errors.New("`data` is no slice type")
	}

	dstType := reflect.TypeOf(dst)
	if dstType.Kind() != reflect.Ptr {
		return errors.New("`dst` is no ptr")
	}
	dstType = dstType.Elem()
	if dstType.Kind() != reflect.Slice {
		return errors.New("`dst` is no slice ptr")
	}

	if dataType.Kind() != dstType.Kind() {
		return errors.New("`data` is not the same type as `dst`")
	}

	var mapType = reflect.MapOf(dataType.Elem(), reflect.TypeOf(reflect.Interface))

	var mapValue = reflect.MakeMap(mapType)
	var dataValue = reflect.Indirect(reflect.ValueOf(data))

	//v := struct{}{}
	for i := 0; i < dataValue.Len(); i++ {
		mapValue.SetMapIndex(dataValue.Index(i), reflect.ValueOf(reflect.Struct))
	}

	dstValue := reflect.Indirect(reflect.ValueOf(dst))
	newDstValue := reflect.MakeSlice(dstType, mapValue.Len(), mapValue.Len())

	//dstValue = dstValue.
	var i = 0
	itr := mapValue.MapRange()
	for itr.Next() {
		newDstValue.Index(i).Set(itr.Key())
		i++
	}
	dstValue.Set(newDstValue)
	return nil
}
