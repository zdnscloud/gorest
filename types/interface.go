package types

type Object interface {
	ObjectID
	ObjectType
	ObjectParent
}

type ObjectParent interface {
	GetParent() Parent
	SetParent(Parent)
}

type ObjectID interface {
	GetID() string
	SetID(string)
}

type ObjectType interface {
	GetType() string
	SetType(string)
}

type Handler interface {
	Create(Object) (interface{}, *APIError)
	Delete(Object) *APIError
	Update(Object) (interface{}, *APIError)
	List(Object) interface{}
	Get(Object) interface{}
	Action(Object, string, map[string]interface{}) (interface{}, *APIError)
}

type AccessControl interface {
	CanCreate(apiContext *APIContext, schema *Schema) *APIError
	CanList(apiContext *APIContext, schema *Schema) *APIError
	CanGet(apiContext *APIContext, schema *Schema) *APIError
	CanUpdate(apiContext *APIContext, schema *Schema) *APIError
	CanDelete(apiContext *APIContext, schema *Schema) *APIError
}
