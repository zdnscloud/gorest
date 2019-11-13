package resourcedoc

import (
	"encoding/json"
	"os"
	"path"
	"reflect"
	"strings"

	slice "github.com/zdnscloud/cement/slice"
	"github.com/zdnscloud/gorest/resource"
	"github.com/zdnscloud/gorest/util"
)

const (
	requiredTag    = "required"
	optionsTag     = "options="
	descriptionTag = "description="
	docFileSuffix  = ".json"
	ignoreType     = "ResourceBase"
	ignoreJsonFlag = "inline"
	ignoreJsonName = "-"
)

type Resource struct {
	ResourceType      string                    `json:"resourceType,omitempty"`
	CollectionName    string                    `json:"collectionName,omitempty"`
	ParentResources   []string                  `json:"parentResources,omitempty"`
	ResourceFields    ResourceFields            `json:"resourceFields,omitempty"`
	SubResources      map[string]ResourceFields `json:"subResources,omitempty"`
	ResourceMethods   []resource.HttpMethod     `json:"resourceMethods,omitempty"`
	CollectionMethods []resource.HttpMethod     `json:"collectionMethods,omitempty"`
}

type ResourceFields map[string]ResourceField

type ResourceField struct {
	Type        string   `json:"type,omitempty"`
	ValidValues []string `json:"validValues,omitempty"`
	ElemType    string   `json:"elemType,omitempty"`
	KeyType     string   `json:"keyType,omitempty"`
	ValueType   string   `json:"valueType,omitempty"`
	Description []string `json:"description,omitempty"`
}

func NewResource(name string, kind resource.ResourceKind, handler resource.Handler, parents []string) *Resource {
	resource := &Resource{
		ResourceType:      name,
		CollectionName:    util.GuessPluralName(name),
		ParentResources:   parents,
		SubResources:      make(map[string]ResourceFields),
		ResourceMethods:   resource.GetResourceMethods(handler),
		CollectionMethods: resource.GetCollectionMethods(handler),
	}
	resource.ResourceFields = buildResourceFields(resource, reflect.TypeOf(kind))
	return resource
}

func (r *Resource) WriteJsonFile(targetPath string) error {
	if err := os.MkdirAll(targetPath, os.ModePerm); err != nil {
		return err
	}
	filePtr, err := os.Create(path.Join(targetPath, r.ResourceType+docFileSuffix))
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	_, err = filePtr.Write(data)
	return err
}

func buildResourceFields(resource *Resource, t reflect.Type) ResourceFields {
	resourceFields := make(map[string]ResourceField)
	for i := 0; i < t.NumField(); i++ {
		name := t.Field(i).Name
		typ := t.Field(i).Type
		tag := t.Field(i).Tag
		jsonName := fieldJsonName(name, tag)
		if (strings.HasSuffix(name, ignoreType) && slice.SliceIndex(strings.Split(tag.Get("json"), ","), ignoreJsonFlag) >= 0) || jsonName == ignoreJsonName {
			continue
		}

		resourceFields[jsonName] = buildResourceField(typ, tag)

		if _, ignore := getTypeIfIgnore(typ.Name()); !ignore {
			if t := getStructType(typ); t != nil {
				resource.SubResources[LowerFirstCharacter(t.Name())] = buildResourceFields(resource, t)
			}
		}
	}
	return resourceFields
}

func buildResourceField(t reflect.Type, tag reflect.StructTag) ResourceField {
	var elemType, keyType, valueType string
	valueRange := parseTag(tag, true)
	typ, ignore := getTypeIfIgnore(t.Name())
	if !ignore {
		if len(valueRange) > 0 {
			typ = Enum
		} else {
			typ = getType(t)
			switch typ {
			case Array:
				elemType = getElemType(t)
			case Map:
				keyType, valueType = getMapElemType(t)
			}
		}
	}
	return ResourceField{
		Type:        typ,
		ElemType:    elemType,
		ValidValues: valueRange,
		KeyType:     keyType,
		ValueType:   valueType,
		Description: parseTag(tag, false),
	}
}

func parseTag(tag reflect.StructTag, isOptions bool) []string {
	var tags []string
	restTags := strings.Split(tag.Get("rest"), ",")
	for _, t := range restTags {
		if isOptions {
			if strings.HasPrefix(t, optionsTag) {
				tags = append(tags, strings.Split(strings.TrimPrefix(t, optionsTag), "|")...)
				break
			}
		} else {
			if strings.HasPrefix(t, requiredTag) {
				tags = append(tags, requiredTag)
			}
			if strings.HasPrefix(t, descriptionTag) {
				tags = append(tags, strings.TrimPrefix(t, descriptionTag))
			}
		}
	}
	return tags
}
