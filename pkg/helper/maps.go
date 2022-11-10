package helper

import (
	"reflect"
)

var (
	MaxDepth = 16
)

// TODO: update this when moving to go 1.18 to use maps Copy
// https://pkg.go.dev/golang.org/x/exp/maps#Copy
func Merge(dst, src map[string]interface{}) map[string]interface{} {
	return merge(dst, src, 0)
}

func merge(dst, src map[string]interface{}, depth int) map[string]interface{} {
	if depth > MaxDepth {
		panic("merge max depth reached !")
	}
	for key, srcVal := range src {
		if dstVal, ok := dst[key]; ok {
			srcMap, isSrcMap := toMap(srcVal)
			dstMap, isDstMap := toMap(dstVal)
			if isSrcMap && isDstMap {
				srcVal = merge(dstMap, srcMap, depth+1)
			}
		}
		dst[key] = srcVal
	}
	return dst
}

func toMap(i interface{}) (map[string]interface{}, bool) {
	value := reflect.ValueOf(i)
	if value.Kind() == reflect.Map {
		m := map[string]interface{}{}
		for _, k := range value.MapKeys() {
			m[k.String()] = value.MapIndex(k).Interface()
		}
		return m, true
	}
	return map[string]interface{}{}, false
}
