package resourcedoc

import (
	"github.com/zdnscloud/gorest/util"
	"reflect"
	"strings"
)

const (
	String  = "string"
	Array   = "array"
	Bool    = "bool"
	Map     = "map"
	Int     = "int"
	Uint    = "uint"
	Enum    = "enum"
	Default = "default"
	Unknow  = "unknow"
)

func cutSymbolUint(in string) string {
	if in == "int8" || in == "int16" || in == "int32" || in == "int64" {
		return Int
	}
	if in == "uint8" || in == "uint16" || in == "uint32" || in == "uint64" {
		return Uint
	}
	return in
}

func setSlice(t reflect.Type) string {
	k := util.Inspect(t)
	switch k {
	case util.StringSlice:
		return String
	case util.IntSlice, util.UintSlice, util.StructSlice, util.StructPtrSlice, util.BoolSlice:
		nestType := t.Elem()
		if k == util.StructPtrSlice {
			nestType = nestType.Elem()
		}
		return cutSymbolUint(nestType.Name())
	}
	return Unknow
}

func setMap(t reflect.Type) (string, string) {
	k := util.Inspect(t)
	switch k {
	case util.StringIntMap, util.StringStringMap, util.StringUintMap, util.StringStructMap, util.StringStructPtrMap:
		nestType := t.Elem()
		if k == util.StringStructPtrMap {
			nestType = nestType.Elem()
		}
		return String, cutSymbolUint(nestType.Name())
	}
	return Unknow, Unknow
}

func setType(t reflect.Type) string {
	k := util.Inspect(t)
	switch k {
	case util.String:
		return String
	case util.Int:
		return Int
	case util.Uint:
		return Uint
	case util.Bool:
		return Bool
	case util.StringIntMap, util.StringStringMap, util.StringUintMap, util.StringStructMap, util.StringStructPtrMap:
		return Map
	case util.IntSlice, util.UintSlice, util.BoolSlice, util.StringSlice, util.StructSlice, util.StructPtrSlice:
		return Array
	case util.Struct:
		return t.Name()
	case util.StructPtr:
		return t.Elem().Name()
	}
	return Unknow
}

func strFirstToLower(str string) string {
	if len(str) < 1 {
		return ""
	}
	strArry := []rune(str)
	if strArry[0] >= 65 && strArry[0] <= 96 {
		strArry[0] += 32
	}
	return string(strArry)
}

func fieldJsonName(name, jsonTag string) string {
	if jsonTag != "" {
		tags := strings.Split(jsonTag, ",")
		for _, tag := range tags {
			if tag != "omitempty" {
				return tag
			}
		}
	}

	return name
}
