package types

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/zdnscloud/gorest/util"
)

func (s *Schemas) MustImport(version *APIVersion, obj ResourceType, objHandler interface{}) *Schemas {
	if reflect.ValueOf(obj).Kind() == reflect.Ptr {
		panic(fmt.Errorf("obj cannot be a pointer"))
	}

	objType := reflect.TypeOf(obj)
	if _, ok := reflect.New(objType).Interface().(Object); ok == false {
		panic("resource type doesn't implement object interface")
	}

	schema, err := s.importType(version, objType)
	if err != nil {
		panic(err)
	}

	handler, err := NewHandler(objHandler)
	if err != nil {
		panic(err)
	}

	schema.Handler = handler
	schema.ResourceMethods = GetResourceMethods(handler)
	schema.CollectionMethods = GetCollectionMethods(handler)
	schema.ResourceActions = obj.GetActions()
	schema.CollectionActions = obj.GetCollectionActions()
	schema.Parents = obj.GetParents()

	return s
}

func (s *Schemas) importType(version *APIVersion, t reflect.Type) (*Schema, error) {
	typeName := s.getTypeName(t)
	existing := s.Schema(version, typeName)
	if existing != nil {
		return existing, nil
	}

	schema, err := s.newSchemaFromType(version, t)
	if err != nil {
		return nil, err
	}

	if _, err := s.AddSchema(schema); err != nil {
		return nil, err
	}

	return s.Schema(&schema.Version, schema.GetType()), nil
}

func (s *Schemas) newSchemaFromType(version *APIVersion, t reflect.Type) (*Schema, error) {
	schema := &Schema{
		Version:        *version,
		ResourceFields: map[string]Field{},
		StructVal:      reflect.New(t).Elem(),
	}

	if err := s.readFields(schema, t); err != nil {
		return nil, err
	}

	return schema, nil
}

func (s *Schemas) readFields(schema *Schema, t reflect.Type) error {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		fieldJsonName, isAnonymous := util.GetFieldJsonName(field)
		if isAnonymous {
			if err := s.readFields(schema, field.Type); err != nil {
				return err
			}
			continue
		}

		if fieldJsonName == "" {
			continue
		}

		schemaField := Field{
			Create:   true,
			Update:   true,
			Nullable: false,
			CodeName: field.Name,
		}

		fieldType := field.Type
		if fieldType.Kind() == reflect.Ptr {
			schemaField.Nullable = true
			fieldType = fieldType.Elem()
		} else if fieldType.Kind() == reflect.Bool {
			schemaField.Default = false
		} else if fieldType.Kind() == reflect.Int ||
			fieldType.Kind() == reflect.Int32 ||
			fieldType.Kind() == reflect.Int64 {
			schemaField.Default = 0
		}

		if err := applyTag(&field, &schemaField); err != nil {
			return err
		}

		if schemaField.Type == "" {
			inferedType, err := s.determineSchemaType(&schema.Version, fieldType)
			if err != nil {
				return fmt.Errorf("failed inspecting type %s, field %s: %v", t, fieldJsonName, err)
			}
			schemaField.Type = inferedType
		}

		if schemaField.Default != nil {
			switch schemaField.Type {
			case "int":
				n, err := util.ToNumber(schemaField.Default)
				if err != nil {
					return err
				}
				schemaField.Default = n
			case "boolean":
				schemaField.Default = util.ToBool(schemaField.Default)
			case "string":
			default:
				return fmt.Errorf("only int, bool and string support default value")
			}
		}

		schema.ResourceFields[fieldJsonName] = schemaField
	}

	return nil
}

func applyTag(structField *reflect.StructField, field *Field) error {
	for _, part := range strings.Split(structField.Tag.Get("rest"), ",") {
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
			field.Nullable = value != "false"
		case "create":
			field.Create = value != "false"
		case "required":
			field.Required = value == "true"
		case "update":
			field.Update = value != "false"
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
		fallthrough
	case reflect.Uint64:
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
		schema, err := s.importType(version, t)
		if err != nil {
			return "", err
		}
		return schema.GetType(), nil
	default:
		return "", fmt.Errorf("unknown type kind %s", t.Kind())
	}
}

func (s *Schemas) getTypeName(t reflect.Type) string {
	if name, ok := s.typeNames[t]; ok {
		return name
	}

	name := strings.ToLower(t.Name())
	s.typeNames[t] = name
	return name
}
