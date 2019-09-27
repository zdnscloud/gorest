package resourcedoc

import (
	"reflect"
	"strings"
)

const (
	requiredTag = "required="
	optionsTag  = "options="
)

func Mapf(s string) string {
	switch s {
	case requiredTag:
		return "field required"
	}
	return ""
}

func Tets(tag reflect.StructTag) string {
	var describe string
	rest := tag.Get("rest")
	restTags := strings.Split(rest, ",")
	for _, tag := range restTags {
		if strings.HasPrefix(tag, requiredTag) {
			describe += Mapf(requiredTag)
		}
		if strings.HasPrefix(tag, optionsTag) {
			describe += Mapf(optionsTag)
		}
	}
	return ""
}

func OptionsTag(tag reflect.StructTag) []string {
	rest := tag.Get("rest")
	restTags := strings.Split(rest, ",")
	for _, t := range restTags {
		if !strings.HasPrefix(t, optionsTag) {
			continue
		}
		requiredVal := strings.TrimPrefix(t, optionsTag)
		return strings.Split(requiredVal, "|")
	}
	return []string{}
}
