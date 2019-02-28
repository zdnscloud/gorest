package types

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/zdnscloud/gorest/types/convert"
)

var (
	namespacedType = reflect.TypeOf(Namespaced{})
	resourceType   = reflect.TypeOf(Resource{})
	blacklistNames = map[string]bool{
		"actions": true,
	}
)

func (s *Schemas) TypeName(name string, obj interface{}) *Schemas {
	s.typeNames[reflect.TypeOf(obj)] = name
	return s
}

func (s *Schemas) getTypeName(t reflect.Type) string {
	if name, ok := s.typeNames[t]; ok {
		return name
	}
	return convert.LowerTitle(t.Name())
}

func (s *Schemas) AddMapperForType(version *APIVersion, obj interface{}, mapper ...Mapper) *Schemas {
	if len(mapper) == 0 {
		return s
	}

	t := reflect.TypeOf(obj)
	typeName := s.getTypeName(t)
	if len(mapper) == 1 {
		return s.AddMapper(version, typeName, mapper[0])
	}
	return s.AddMapper(version, typeName, Mappers(mapper))
}

func (s *Schemas) MustImport(version *APIVersion, obj interface{}, externalOverrides ...interface{}) *Schemas {
	if reflect.ValueOf(obj).Kind() == reflect.Ptr {
		panic(fmt.Errorf("obj cannot be a pointer"))
	}

	if _, err := s.Import(version, obj, externalOverrides...); err != nil {
		panic(err)
	}
	return s
}

func (s *Schemas) MustImportAndCustomize(version *APIVersion, obj interface{}, handler Handler, f func(*Schema, Handler), externalOverrides ...interface{}) *Schemas {
	return s.MustImport(version, obj, externalOverrides...).
		MustCustomizeType(version, obj, handler, f)
}

func (s *Schemas) Import(version *APIVersion, obj interface{}, externalOverrides ...interface{}) (*Schema, error) {
	var types []reflect.Type
	for _, override := range externalOverrides {
		types = append(types, reflect.TypeOf(override))
	}

	return s.importType(version, reflect.TypeOf(obj), types...)
}

func (s *Schemas) newSchemaFromType(version *APIVersion, t reflect.Type, typeName string) (*Schema, error) {
	schema := &Schema{
		ID:                typeName,
		Version:           *version,
		CodeName:          t.Name(),
		PkgName:           t.PkgPath(),
		ResourceFields:    map[string]Field{},
		ResourceActions:   map[string]Action{},
		CollectionActions: map[string]Action{},
		StructVal:         reflect.New(t).Elem(),
	}

	if err := s.readFields(schema, t); err != nil {
		return nil, err
	}

	return schema, nil
}

func (s *Schemas) MustCustomizeType(version *APIVersion, obj interface{}, handler Handler, f func(*Schema, Handler)) *Schemas {
	name := s.getTypeName(reflect.TypeOf(obj))
	schema := s.Schema(version, name)
	if schema == nil {
		panic("Failed to find schema " + name)
	}

	f(schema, handler)

	return s
}

func (s *Schemas) importType(version *APIVersion, t reflect.Type, overrides ...reflect.Type) (*Schema, error) {
	typeName := s.getTypeName(t)
	existing := s.Schema(version, typeName)
	if existing != nil {
		return existing, nil
	}

	schema, err := s.newSchemaFromType(version, t, typeName)
	if err != nil {
		return nil, err
	}

	mappers := s.mapper(&schema.Version, schema.ID)
	if s.DefaultMappers != nil {
		if schema.CanList(nil) == nil {
			mappers = append(s.DefaultMappers(), mappers...)
		}
	}
	if s.DefaultPostMappers != nil {
		mappers = append(mappers, s.DefaultPostMappers()...)
	}

	if len(mappers) > 0 {
		copy, err := s.newSchemaFromType(version, t, typeName)
		if err != nil {
			return nil, err
		}
		schema.InternalSchema = copy
	}

	for _, override := range overrides {
		if err := s.readFields(schema, override); err != nil {
			return nil, err
		}
	}

	mapper := &typeMapper{
		Mappers: mappers,
		root:    schema.CanList(nil) == nil,
	}

	if err := mapper.ModifySchema(schema, s); err != nil {
		return nil, err
	}

	schema.Mapper = mapper
	s.AddSchema(*schema)

	return s.Schema(&schema.Version, schema.ID), s.Err()
}

func jsonName(f reflect.StructField) string {
	return strings.SplitN(f.Tag.Get("json"), ",", 2)[0]
}

func (s *Schemas) readFields(schema *Schema, t reflect.Type) error {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		if field.PkgPath != "" {
			// unexported field
			continue
		}

		jsonName := jsonName(field)
		if jsonName == "-" {
			continue
		}

		if field.Anonymous && jsonName == "" {
			t := field.Type
			if t.Kind() == reflect.Ptr {
				t = t.Elem()
			}
			if t.Kind() == reflect.Struct {
				if t == namespacedType {
					schema.Scope = NamespaceScope
				}
				if err := s.readFields(schema, t); err != nil {
					return err
				}
			}
			continue
		}

		fieldName := jsonName
		if fieldName == "" {
			fieldName = convert.LowerTitle(field.Name)
			if strings.HasSuffix(fieldName, "ID") {
				fieldName = strings.TrimSuffix(fieldName, "ID") + "Id"
			}
		}

		if blacklistNames[fieldName] {
			continue
		}

		schemaField := Field{
			Create:   true,
			Update:   true,
			Nullable: true,
			CodeName: field.Name,
		}

		fieldType := field.Type
		if fieldType.Kind() == reflect.Ptr {
			schemaField.Nullable = true
			fieldType = fieldType.Elem()
		} else if fieldType.Kind() == reflect.Bool {
			schemaField.Nullable = false
			schemaField.Default = false
		} else if fieldType.Kind() == reflect.Int ||
			fieldType.Kind() == reflect.Int32 ||
			fieldType.Kind() == reflect.Int64 {
			schemaField.Nullable = false
			schemaField.Default = 0
		}

		if err := applyTag(&field, &schemaField); err != nil {
			return err
		}

		if schemaField.Type == "" {
			inferedType, err := s.determineSchemaType(&schema.Version, fieldType)
			if err != nil {
				return fmt.Errorf("failed inspecting type %s, field %s: %v", t, fieldName, err)
			}
			schemaField.Type = inferedType
		}

		if schemaField.Default != nil {
			switch schemaField.Type {
			case "int":
				n, err := convert.ToNumber(schemaField.Default)
				if err != nil {
					return err
				}
				schemaField.Default = n
			case "boolean":
				schemaField.Default = convert.ToBool(schemaField.Default)
			}
		}

		schema.ResourceFields[fieldName] = schemaField
	}

	return nil
}

func applyTag(structField *reflect.StructField, field *Field) error {
	for _, part := range strings.Split(structField.Tag.Get("singlecloud"), ",") {
		if part == "" {
			continue
		}

		var err error
		key, value := getKeyValue(part)

		switch key {
		case "type":
			field.Type = value
		case "codeName":
			field.CodeName = value
		case "default":
			field.Default = value
		case "nullable":
			field.Nullable = true
		case "notnullable":
			field.Nullable = false
		case "nocreate":
			field.Create = false
		case "writeOnly":
			field.WriteOnly = true
		case "required":
			field.Required = true
		case "noupdate":
			field.Update = false
		case "minLength":
			field.MinLength, err = toInt(value, structField)
		case "maxLength":
			field.MaxLength, err = toInt(value, structField)
		case "min":
			field.Min, err = toInt(value, structField)
		case "max":
			field.Max, err = toInt(value, structField)
		case "options":
			field.Options = split(value)
			if field.Type == "" {
				field.Type = "enum"
			}
		case "validChars":
			field.ValidChars = value
		case "invalidChars":
			field.InvalidChars = value
		default:
			return fmt.Errorf("invalid tag %s on field %s", key, structField.Name)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func toInt(value string, structField *reflect.StructField) (*int64, error) {
	i, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid number on field %s: %v", structField.Name, err)
	}
	return &i, nil
}

func split(input string) []string {
	result := []string{}
	for _, i := range strings.Split(input, "|") {
		for _, part := range strings.Split(i, " ") {
			part = strings.TrimSpace(part)
			if len(part) > 0 {
				result = append(result, part)
			}
		}
	}

	return result
}

func getKeyValue(input string) (string, string) {
	var (
		key, value string
	)
	parts := strings.SplitN(input, "=", 2)
	key = parts[0]
	if len(parts) > 1 {
		value = parts[1]
	}

	return key, value
}

func deRef(p reflect.Type) reflect.Type {
	if p.Kind() == reflect.Ptr {
		return p.Elem()
	}
	return p
}

func (s *Schemas) determineSchemaType(version *APIVersion, t reflect.Type) (string, error) {
	switch t.Kind() {
	case reflect.Uint8:
		return "byte", nil
	case reflect.Bool:
		return "boolean", nil
	case reflect.Int:
		fallthrough
	case reflect.Int32:
		fallthrough
	case reflect.Uint32:
		fallthrough
	case reflect.Int64:
		return "int", nil
	case reflect.Interface:
		return "json", nil
	case reflect.Map:
		subType, err := s.determineSchemaType(version, deRef(t.Elem()))
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("map[%s]", subType), nil
	case reflect.Slice:
		subType, err := s.determineSchemaType(version, deRef(t.Elem()))
		if err != nil {
			return "", err
		}
		if subType == "byte" {
			return "base64", nil
		}
		return fmt.Sprintf("array[%s]", subType), nil
	case reflect.String:
		return "string", nil
	case reflect.Struct:
		if t.Name() == "Time" {
			return "date", nil
		}
		if t.Name() == "IntOrString" {
			return "intOrString", nil
		}
		if t.Name() == "Quantity" {
			return "string", nil
		}
		schema, err := s.importType(version, t)
		if err != nil {
			return "", err
		}
		return schema.ID, nil
	default:
		return "", fmt.Errorf("unknown type kind %s", t.Kind())
	}

}
