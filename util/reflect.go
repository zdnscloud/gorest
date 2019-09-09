package util

import (
	"reflect"
	"strings"
)

var (
	blacklistNames = map[string]bool{
		"actions":           true,
		"links":             true,
		"creationTimestamp": true,
	}
)

func GetFieldJsonName(field reflect.StructField) (string, bool) {
	if field.PkgPath != "" {
		return "", false
	}

	jsonName := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
	if jsonName == "-" {
		return "", false
	}

	if field.Anonymous && jsonName == "" {
		t := field.Type
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		return "", t.Kind() == reflect.Struct
	}

	fieldJsonName := jsonName
	if fieldJsonName == "" {
		fieldJsonName = strings.ToLower(field.Name)
		if strings.HasSuffix(fieldJsonName, "ID") {
			fieldJsonName = strings.TrimSuffix(fieldJsonName, "ID") + "Id"
		}
	}

	if blacklistNames[fieldJsonName] {
		return "", false
	}

	return fieldJsonName, false
}
