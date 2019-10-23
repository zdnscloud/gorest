package resourcedoc

import (
	"reflect"
	"strings"
)

const (
	requiredTag    = "required="
	optionsTag     = "options="
	descriptionTag = "description="
	required       = "required"
)

func DescriptionTag(tag reflect.StructTag) []string {
	var describe []string
	rest := tag.Get("rest")
	restTags := strings.Split(rest, ",")
	for _, t := range restTags {
		if strings.HasPrefix(t, requiredTag) {
			describe = append(describe, required)
		}
		if strings.HasPrefix(t, descriptionTag) {
			descriptionVal := strings.TrimPrefix(t, descriptionTag)
			describe = append(describe, descriptionVal)
		}
	}
	return describe
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
