package resourcedoc

import (
	"reflect"
	"strings"

	"github.com/zdnscloud/gorest/util"
)

const (
	Array  = "array"
	Map    = "map"
	Enum   = "enum"
	Unknow = "unknow"
)

func getTypeIfIgnore(typ string) (string, bool) {
	switch typ {
	case "RawMessage":
		return "json", true
	case "ISOTime":
		return "date", true
	default:
		return "", false
	}
}

func getStructType(typ reflect.Type) reflect.Type {
	switch util.Inspect(typ) {
	case util.StringStructPtrMap, util.StructPtrSlice:
		return typ.Elem().Elem()
	case util.StringStructMap, util.StructPtr, util.StructSlice:
		return typ.Elem()
	case util.Struct:
		return typ
	default:
		return nil
	}
}

func getType(t reflect.Type) string {
	switch k := util.Inspect(t); k {
	case util.String, util.Int, util.Uint, util.Bool:
		return string(k)
	case util.StringIntMap, util.StringStringMap, util.StringUintMap, util.StringStructMap, util.StringStructPtrMap:
		return Map
	case util.IntSlice, util.UintSlice, util.BoolSlice, util.StringSlice, util.StructSlice, util.StructPtrSlice:
		return Array
	case util.Struct:
		return LowerFirstCharacter(t.Name())
	case util.StructPtr:
		return LowerFirstCharacter(t.Elem().Name())
	default:
		return Unknow
	}
}

func getElemType(t reflect.Type) string {
	switch k := util.Inspect(t); k {
	case util.StringSlice:
		return string(util.String)
	case util.IntSlice:
		return string(util.Int)
	case util.UintSlice:
		return string(util.Uint)
	case util.StructSlice, util.StructPtrSlice, util.BoolSlice:
		nestType := t.Elem()
		if k == util.StructPtrSlice {
			nestType = nestType.Elem()
		}
		return LowerFirstCharacter(nestType.Name())
	default:
		return Unknow
	}
}

func getMapElemType(t reflect.Type) (string, string) {
	switch k := util.Inspect(t); k {
	case util.StringIntMap:
		return string(util.String), string(util.Int)
	case util.StringUintMap:
		return string(util.String), string(util.Uint)
	case util.StringStringMap, util.StringStructMap, util.StringStructPtrMap:
		nestType := t.Elem()
		if k == util.StringStructPtrMap {
			nestType = nestType.Elem()
		}
		return string(util.String), LowerFirstCharacter(nestType.Name())
	default:
		return Unknow, Unknow

	}
}

func LowerFirstCharacter(original string) string {
	if len(original) > 0 {
		original = strings.ToLower(original[:1]) + original[1:]
	}
	return original
}

func fieldJsonName(name string, tag reflect.StructTag) string {
	if jsonTag := tag.Get("json"); jsonTag != "" {
		tags := strings.Split(jsonTag, ",")
		for _, tag := range tags {
			if tag != "omitempty" {
				return tag
			}
		}
	}
	return name
}
