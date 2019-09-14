package resourcefield

type Field interface {
	JsonName() string

	Name() string

	//default value for field
	//if not set default return nil
	DefaultValue() interface{}
	SetDefault(interface{})

	IsRequired() bool
	SetRequired(bool)

	//validate fields of go struct
	Validate(interface{}) error

	//work on json format string
	CheckRequired(json map[string]interface{}) error
	FillDefault(json map[string]interface{})
}

type Validator interface {
	//validate each field is valid
	Validate(interface{}) error
}
