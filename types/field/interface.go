package field

type Field interface {
	JsonName() string

	Name() string

	//default value for field
	//if not set default return nil
	DefaultValue() interface{}
	SetDefault(interface{})

	IsRequired() bool
	SetRequired(bool)

	Validate(interface{}) error

	CheckRequired(json map[string]interface{}) error
	FillDefault(json map[string]interface{})
}

type Validator interface {
	Validate(interface{}) error
}
