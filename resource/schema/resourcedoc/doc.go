package resourcedoc

import (
	"encoding/json"
	"github.com/zdnscloud/gorest/resource"
	"os"
	"reflect"
)

type Document struct {
	ResourceType      string                             `json:"resourceType,omitempty"`
	CollectionName    string                             `json:"collectionName,omitempty"`
	ResourceFields    []map[string]DocField              `json:"resourceFields,omitempty"`
	SubResources      []map[string][]map[string]DocField `json:"subResources,omitempty"`
	ResourceMethods   []resource.HttpMethod              `json:"resourceMethods,omitempty"`
	CollectionMethods []resource.HttpMethod              `json:"collectionMethods,omitempty"`
}

type DocField struct {
	Type        string   `json:"type,omitempty"`
	ValidValues []string `json:"validValues,omitempty"`
	ElemType    string   `json:"elemType,omitempty"`
	KeyType     string   `json:"keyType,omitempty"`
	ValueType   string   `json:"valueType,omitempty"`
	Describe    string   `json:"describe,omitempty"`
}

type Docer struct {
	resourceName string
	resourceKind resource.ResourceKind
	document     Document
}

func NewDocer(name string, kind resource.ResourceKind, handler resource.Handler) *Docer {
	builder := NewBuilder()
	builder.BuildResource(name, reflect.TypeOf(kind))

	var resourceType, collectionName string
	var resourceFields []map[string]DocField
	for _, v := range builder.GetTop() {
		for _, f := range v {
			field := fieldToDoc(f)
			if len(field) == 0 {
				continue
			}
			resourceFields = append(resourceFields, field)
		}
		resourceType = name
		collectionName = name + "s"
	}
	subresources := make([]map[string][]map[string]DocField, 0)
	for _, resource := range builder.GetSub() {
		for k, v := range resource {
			subresource := make(map[string][]map[string]DocField)
			var fields []map[string]DocField
			for _, f := range v {
				field := fieldToDoc(f)
				if len(field) == 0 {
					continue
				}
				fields = append(fields, field)
			}
			name := strFirstToLower(k)
			subresource[name] = fields
			subresources = append(subresources, subresource)
		}
	}
	return &Docer{
		resourceName: name,
		resourceKind: kind,
		document: Document{
			ResourceType:      resourceType,
			CollectionName:    collectionName,
			ResourceFields:    resourceFields,
			SubResources:      subresources,
			ResourceMethods:   resource.GetResourceMethods(handler),
			CollectionMethods: resource.GetCollectionMethods(handler),
		},
	}
}

func (d *Docer) WriteConfig(path string) error {
	data, err := json.MarshalIndent(d.document, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}
	file := path + "/" + d.document.ResourceType + ".json"
	filePtr, err := os.Create(file)
	if err != nil {
		return err
	}
	filePtr.Write(data)
	return nil
}

func fieldToDoc(f Field) map[string]DocField {
	var et, kt, vt string
	t := setType(f.Type)
	vv := OptionsTag(f.Tag)
	if len(vv) > 0 {
		t = Enum
	}
	if t == Array {
		et = setSlice(f.Type)
	}
	if t == Map {
		kt, vt = setMap(f.Type)
	}
	field := make(map[string]DocField)
	jsonname := fieldJsonName(f.Name, f.Tag.Get("json"))
	newname := strFirstToLower(jsonname)
	field[newname] = DocField{
		Type:        strFirstToLower(t),
		ElemType:    strFirstToLower(et),
		ValidValues: vv,
		KeyType:     strFirstToLower(kt),
		ValueType:   strFirstToLower(vt),
	}
	return field
}
