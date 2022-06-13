/**
 * @Author: zzy
 * @Email: zhangzhongyuan@didiglobal.com
 * @Description:
 * @File: type.go
 * @Package: utils
 * @Version: 1.0.0
 * @Date: 2022/6/13 15:30
 */

package utils

import (
	"reflect"
	"strings"
)

func parseTag(tag string) (name string, options string) {
	if idx := strings.Index(tag, ","); idx != -1 {
		return tag[:idx], tag[idx+1:]
	}
	return tag, ""
}

// StructToMap 结构体转map
func StructToMap(v interface{}, tag string) map[string]interface{} {
	var (
		val    = reflect.ValueOf(v)
		typ    = val.Type()
		fields = make(map[string]interface{})
	)

	for i := 0; i < typ.NumField(); i++ {
		if typ.Field(i).Anonymous {
			for key, val := range StructToMap(val.Field(i).Interface(), tag) {
				fields[key] = val
			}
			continue
		}

		if tag == "" {
			fields[typ.Field(i).Name] = val.Field(i).Interface()
			continue
		}

		tag := typ.Field(i).Tag.Get(tag)
		name, options := parseTag(tag)
		if val.Field(i).IsZero() && strings.Contains(options, "omitempty") {
			continue
		}

		fields[name] = val.Field(i).Interface()
	}

	return fields
}
