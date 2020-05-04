package db

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/zdnscloud/cement/reflector"
	"github.com/zdnscloud/cement/stringtool"

	"github.com/zdnscloud/gorest/resource"
)

type Datatype int

const (
	String Datatype = iota
	Int
	Uint32
	Time
	IntArray
	StringArray
	Bool
)

const EmbedResource string = "ResourceBase"
const DBTag string = "db"

type Check string

const (
	NoCheck  Check = ""
	Positive Check = "positive"
)

type ResourceField struct {
	Name   string
	Type   Datatype
	Unique bool
	Check  Check
}

type ResourceDescriptor struct {
	Typ            ResourceType
	Fields         []ResourceField
	Pks            []ResourceType
	Uks            []ResourceType
	Owners         []ResourceType
	Refers         []ResourceType
	IsRelationship bool
}

type ResourceRelationship struct {
	Typ   ResourceType
	Owner ResourceType
	Refer ResourceType
}

type ResourceMeta struct {
	resources   []ResourceType //resources has dependencies, resources to store their order
	descriptors map[ResourceType]*ResourceDescriptor
	goTypes     map[ResourceType]reflect.Type
}

func NewResourceMeta(resources []resource.Resource) (*ResourceMeta, error) {
	meta := &ResourceMeta{
		resources:   []ResourceType{},
		descriptors: make(map[ResourceType]*ResourceDescriptor),
		goTypes:     make(map[ResourceType]reflect.Type),
	}

	for _, r := range resources {
		if err := meta.Register(r); err != nil {
			return nil, err
		}
	}
	return meta, nil
}

func (meta *ResourceMeta) Clear() {
	for _, r := range meta.resources {
		delete(meta.descriptors, r)
		delete(meta.goTypes, r)
	}
}

func (meta *ResourceMeta) Has(typ ResourceType) bool {
	return meta.descriptors[typ] != nil
}

func (meta *ResourceMeta) GetGoType(typ ResourceType) (reflect.Type, error) {
	if gtyp, ok := meta.goTypes[typ]; !ok {
		return nil, fmt.Errorf("model %v is unknown", typ)
	} else {
		return gtyp, nil
	}
}

func (meta *ResourceMeta) Register(r resource.Resource) error {
	typ := ResourceDBType(r)
	if meta.Has(typ) {
		return fmt.Errorf("duplicate model:%v", typ)
	}

	descriptor, err := genDescriptor(r)
	if err != nil {
		return err
	}

	for _, m := range append(descriptor.Owners, descriptor.Refers...) {
		_, ok := meta.descriptors[m]
		if ok == false {
			return fmt.Errorf("model %v refer to %v is unknown", typ, m)
		}
	}

	meta.resources = append(meta.resources, typ)
	meta.descriptors[typ] = descriptor
	meta.goTypes[typ] = reflect.TypeOf(r).Elem()
	return nil
}

func parseField(name string, typ *reflect.Type) (*ResourceField, error) {
	kind := (*typ).Kind()
	switch kind {
	case reflect.Int:
		return &ResourceField{Name: name, Type: Int}, nil
	case reflect.Uint32:
		return &ResourceField{Name: name, Type: Uint32}, nil
	case reflect.String:
		return &ResourceField{Name: name, Type: String}, nil
	case reflect.Bool:
		return &ResourceField{Name: name, Type: Bool}, nil
	case reflect.Struct:
		if (*typ).String() == "time.Time" {
			return &ResourceField{Name: name, Type: Time}, nil
		} else {
			return nil, fmt.Errorf("model field type is unsupported struct:%v", kind)
		}
	case reflect.Array, reflect.Slice:
		elemKind := (*typ).Elem().Kind()
		if elemKind == reflect.Int {
			return &ResourceField{Name: name, Type: IntArray}, nil
		} else if elemKind == reflect.String {
			return &ResourceField{Name: name, Type: StringArray}, nil
		} else {
			return nil, fmt.Errorf("model field type [%v] is unsupported", elemKind.String())
		}
	default:
		return nil, fmt.Errorf("model field type %v is unsupported", kind.String())
	}
}

func genDescriptor(r resource.Resource) (*ResourceDescriptor, error) {
	fields := []ResourceField{
		ResourceField{Name: IDField, Type: String},
		ResourceField{Name: CreateTimeField, Type: Time},
	}
	pks := []ResourceType{IDField}
	uks := []ResourceType{}
	owners := []ResourceType{}
	refers := []ResourceType{}

	v, ok := reflector.GetStructFromPointer(r)
	if ok == false {
		return nil, fmt.Errorf("need structure pointer but get %v", v.Kind().String())
	}

	rtype := v.Type()
	typ := ResourceDBType(r)

	for i := 0; i < rtype.NumField(); i++ {
		field := rtype.Field(i)
		if field.Name == EmbedResource {
			continue
		}

		oFieldName := field.Name
		fieldName := stringtool.ToSnake(oFieldName)
		fieldTag := field.Tag.Get(DBTag)
		if tagContains(fieldTag, "-") {
			continue
		}

		if fieldName == IDField || fieldName == CreateTimeField {
			return nil, fmt.Errorf("has duplicate id or createTime field which already exists in resource base")
		}

		if tagContains(fieldTag, "ownby") {
			owners = append(owners, ResourceType(fieldName))
		} else if tagContains(fieldTag, "referto") {
			refers = append(refers, ResourceType(fieldName))
		} else {
			newfield, err := parseField(fieldName, &field.Type)
			if err == nil {
				if tagContains(fieldTag, "suk") {
					newfield.Unique = true
				} else {
					newfield.Unique = false
				}

				if tagContains(fieldTag, "positive") {
					newfield.Check = Positive
				}
				fields = append(fields, *newfield)
			} else {
				fmt.Printf("!!!! warning, field %s parse failed %s\n", fieldName, err.Error())
			}
		}

		if tagContains(fieldTag, "pk") {
			pks = append(pks, ResourceType(fieldName))
		} else if tagContains(fieldTag, "uk") {
			uks = append(uks, ResourceType(fieldName))
		}
	}

	isRelationship := len(fields) == 1 && len(owners) == 1 && len(refers) == 1

	return &ResourceDescriptor{
		Typ:            typ,
		Fields:         fields,
		Pks:            pks,
		Uks:            uks,
		Owners:         owners,
		Refers:         refers,
		IsRelationship: isRelationship,
	}, nil
}

func (meta *ResourceMeta) GetDescriptor(typ ResourceType) (*ResourceDescriptor, error) {
	if meta.Has(typ) {
		return meta.descriptors[typ], nil
	} else {
		return nil, fmt.Errorf("model %v is unknown", typ)
	}
}

func (meta *ResourceMeta) GetDescriptors() []*ResourceDescriptor {
	descriptors := []*ResourceDescriptor{}
	for _, r := range meta.resources {
		descriptors = append(descriptors, meta.descriptors[r])
	}
	return descriptors
}

func (meta *ResourceMeta) Resources() []ResourceType {
	return meta.resources
}

//borrow from encoding/json/tags.go
func tagContains(o string, optionName string) bool {
	if len(o) == 0 {
		return false
	}
	s := string(o)
	for s != "" {
		var next string
		i := strings.Index(s, ",")
		if i >= 0 {
			s, next = s[:i], s[i+1:]
		}
		if s == optionName {
			return true
		}
		s = next
	}
	return false
}

func (descriptor *ResourceDescriptor) GetRelationship() *ResourceRelationship {
	if descriptor.IsRelationship == true {
		return &ResourceRelationship{descriptor.Typ, descriptor.Owners[0], descriptor.Refers[0]}
	} else {
		return nil
	}
}

func ResourceToMap(r resource.Resource) (map[string]interface{}, error) {
	v, ok := reflector.GetStructFromPointer(r)
	if ok == false {
		return nil, fmt.Errorf("need structure pointer but get %v", v.Kind().String())
	}

	m := make(map[string]interface{})
	typ := v.Type()
	for i := 0; i < v.NumField(); i++ {
		f := typ.Field(i)
		n := f.Name
		if n == EmbedResource {
			continue
		}

		if tagContains(f.Tag.Get(DBTag), "-") {
			continue
		}

		n = stringtool.ToSnake(n)
		if n == "id" {
			continue
		}
		m[n] = v.Field(i).Interface()
	}
	return m, nil
}
