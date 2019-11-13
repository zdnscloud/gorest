package resourcedoc

import (
	"encoding/json"
	"os"
	"path"
	"reflect"
	"strings"

	"github.com/zdnscloud/gorest/resource"
	"github.com/zdnscloud/gorest/util"
)

const (
	requiredTag    = "required"
	optionsTag     = "options="
	descriptionTag = "description="
)

type ResourceInfo struct {
	ResourceType      string                       `json:"resourceType,omitempty"`
	CollectionName    string                       `json:"collectionName,omitempty"`
	ParentResources   []string                     `json:"parentResources,omitempty"`
	ResourceFields    resourceJsonField            `json:"resourceFields,omitempty"`
	SubResources      map[string]resourceJsonField `json:"subResources,omitempty"`
	ResourceMethods   []resource.HttpMethod        `json:"resourceMethods,omitempty"`
	CollectionMethods []resource.HttpMethod        `json:"collectionMethods,omitempty"`
}

type resourceJsonField map[string]JsonField

type JsonField struct {
	Type        string   `json:"type,omitempty"`
	ValidValues []string `json:"validValues,omitempty"`
	ElemType    string   `json:"elemType,omitempty"`
	KeyType     string   `json:"keyType,omitempty"`
	ValueType   string   `json:"valueType,omitempty"`
	Description []string `json:"description,omitempty"`
}

func NewResourceInfo(name string, kind resource.ResourceKind, handler resource.Handler, parents []string) *ResourceInfo {
	res := NewResource(reflect.TypeOf(kind))

	resourceFields := genResourceJsonField(res.ResourceField)
	subResources := make(map[string]resourceJsonField)
	for name, resource := range res.SubResourceField {
		subResources[name] = genResourceJsonField(resource)
	}
	return &ResourceInfo{
		ResourceType:      name,
		CollectionName:    util.GuessPluralName(name),
		ParentResources:   parents,
		ResourceFields:    resourceFields,
		SubResources:      subResources,
		ResourceMethods:   resource.GetResourceMethods(handler),
		CollectionMethods: resource.GetCollectionMethods(handler),
	}
}

func (d *ResourceInfo) WriteJsonFile(targetPath string) error {
	if err := os.MkdirAll(targetPath, os.ModePerm); err != nil {
		return err
	}
	filePtr, err := os.Create(path.Join(targetPath, d.ResourceType+".json"))
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return err
	}
	_, err = filePtr.Write(data)
	return err
}

func genResourceJsonField(resourceField map[string]Field) map[string]JsonField {
	resourceJsonField := make(map[string]JsonField)
	for name, field := range resourceField {
		resourceJsonField[name] = fieldToJsonField(field)
	}
	return resourceJsonField
}

func fieldToJsonField(f Field) JsonField {
	var typ, elemType, keyType, valueType string
	valueRange := parseTag(f.Tag, true)
	typ, ignore := getTypeIfIgnore(f.Type.Name())
	if !ignore {
		if len(valueRange) > 0 {
			typ = Enum
		} else {
			typ = getType(f.Type)
			switch typ {
			case Array:
				elemType = getElemType(f.Type)
			case Map:
				keyType, valueType = getMapElemType(f.Type)
			}
		}
	}
	return JsonField{
		Type:        typ,
		ElemType:    elemType,
		ValidValues: valueRange,
		KeyType:     keyType,
		ValueType:   valueType,
		Description: parseTag(f.Tag, false),
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
