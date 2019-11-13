package resourcedoc

import (
	"reflect"
	"strings"

	slice "github.com/zdnscloud/cement/slice"
)

type Resource struct {
	ResourceField    resourceField
	SubResourceField map[string]resourceField
}

type resourceField map[string]Field

type Field struct {
	Type reflect.Type
	Tag  reflect.StructTag
}

func NewResource(t reflect.Type) *Resource {
	resource := Resource{
		ResourceField:    make(map[string]Field),
		SubResourceField: make(map[string]resourceField),
	}
	resource.ResourceField = genResourceField(&resource, t)
	return &resource
}

func genResourceField(resource *Resource, t reflect.Type) resourceField {
	resourceField := make(map[string]Field)
	for i := 0; i < t.NumField(); i++ {
		name := t.Field(i).Name
		typ := t.Field(i).Type
		tag := t.Field(i).Tag
		if (strings.HasSuffix(name, "ResourceBase") && slice.SliceIndex(strings.Split(tag.Get("json"), ","), "inline") >= 0) || fieldJsonName(name, tag) == "-" {
			continue
		}

		resourceField[fieldJsonName(name, tag)] = Field{
			Type: typ,
			Tag:  tag,
		}

		if _, ignore := getTypeIfIgnore(typ.Name()); !ignore {
			if t := getStructType(typ); t != nil {
				resource.SubResourceField[LowerFirstCharacter(t.Name())] = genResourceField(resource, t)
			}
		}
	}
	return resourceField
}
