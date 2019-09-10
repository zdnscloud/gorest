package field

import (
	"encoding/json"
	"fmt"
	ut "github.com/zdnscloud/cement/unittest"
	"reflect"
	"testing"
)

func TestFieldBuild(t *testing.T) {
	builder := NewBuilder()
	sf, err := builder.Build(reflect.TypeOf(TestStruct{}))
	ut.Assert(t, err == nil, "")

	ut.Equal(t, len(sf.fields), 12)
	fieldNames := []string{
		"Id",
		"Age",
		"Name",
		"StringWithOption",
		"StringWithDefault",
		"StringWithLenLimit",
		"IntWithDefault",
		"IntWithRange",
		"BoolWithDefault",
		"Composition",
	}
	for _, name := range fieldNames {
		_, ok := sf.fields[name]
		ut.Assert(t, ok, "")
	}
}

func TestInvalidField(t *testing.T) {
	type S1 struct {
		StringWithOption string `json:"stringWithOption,omitempty" rest:"required=true,default=xxxx"`
	}
	builder := NewBuilder()
	_, err := builder.Build(reflect.TypeOf(S1{}))
	ut.Assert(t, err != nil, "")

	type S2 struct {
		StringWithLenLimit string `json:"stringWithLenLimit" rest:"minLen=20,maxLen=10"`
	}
	builder = NewBuilder()
	_, err = builder.Build(reflect.TypeOf(S2{}))
	ut.Assert(t, err != nil, "")

	type S3 struct {
		IntWithDefault int `json:"intWithDefault" rest:"default=boy"`
	}
	builder = NewBuilder()
	_, err = builder.Build(reflect.TypeOf(S3{}))
	ut.Assert(t, err != nil, "")

	type S4 struct {
		IntWithRange uint32 `json:"intWithRange" rest:"min=100,max=10"`
	}
	builder = NewBuilder()
	_, err = builder.Build(reflect.TypeOf(S4{}))
	ut.Assert(t, err != nil, "")

	type S5 struct {
		BoolWithDefault bool `json:"boolWithDefault" rest:"default=fuck"`
	}
	builder = NewBuilder()
	_, err = builder.Build(reflect.TypeOf(S5{}))
	ut.Assert(t, err != nil, "")
}

func TestFieldFillDefault(t *testing.T) {
	builder := NewBuilder()
	sf, _ := builder.Build(reflect.TypeOf(TestStruct{}))

	ts := TestStruct{
		Name:               "a",
		StringWithOption:   "ceph",
		StringWithLenLimit: "aaa",
		IntWithRange:       100,
		Composition: []IncludeStruct{
			IncludeStruct{
				Int8WithRange: 5,
			},
		},
	}
	ts.StringMapCompostion = make(map[string]IncludeStruct)
	ts.StringMapCompostion["a"] = IncludeStruct{
		Int8WithRange: 10,
	}

	ts.IntMapCompostion = make(map[int32]IncludeStruct)
	ts.IntMapCompostion[20] = IncludeStruct{
		Int8WithRange: 20,
	}

	raw := make(map[string]interface{})
	rawByte, _ := json.Marshal(ts)
	json.Unmarshal(rawByte, &raw)
	delete(raw, "Id")
	delete(raw, "Age")
	delete(raw, "stringWithDefault")
	delete(raw, "intWithDefault")
	delete(raw, "boolWithDefault")
	sf.FillDefault(raw)

	rawByte, err := json.Marshal(raw)
	ut.Assert(t, err == nil, "marshal get err %v", err)
	json.Unmarshal(rawByte, &ts)
	ut.Equal(t, ts.Id, "xxxx")
	ut.Equal(t, ts.Age, int64(20))
	ut.Equal(t, ts.StringWithDefault, "boy")
	ut.Equal(t, ts.IntWithDefault, 11)
	ut.Equal(t, ts.BoolWithDefault, true)
	ut.Equal(t, ts.IntMapCompostion[20].Uint16WithDefault, uint16(11))
	ut.Equal(t, ts.StringMapCompostion["a"].Uint16WithDefault, uint16(11))
}

func TestCheckRequired(t *testing.T) {
	builder := NewBuilder()
	sf, _ := builder.Build(reflect.TypeOf(TestStruct{}))
	ts := TestStruct{
		Name:               "dd",
		StringWithOption:   "ceph",
		StringWithLenLimit: "aaa",
		IntWithRange:       100,
		Composition: []IncludeStruct{
			IncludeStruct{
				Int8WithRange: 5,
			},
		},
	}

	raw := make(map[string]interface{})
	rawByte, _ := json.Marshal(ts)

	fmt.Printf("--> rawByte:%s\n", string(rawByte))
	json.Unmarshal(rawByte, &raw)
	ut.Assert(t, sf.CheckRequired(raw) == nil, "")

	for _, name := range []string{"name", "stringWithOption", "stringMapComposition", "intMapComposition"} {
		json.Unmarshal(rawByte, &raw)
		delete(raw, name)
		ut.Assert(t, sf.CheckRequired(raw) != nil, "")
	}
}

func TestValidate(t *testing.T) {
	builder := NewBuilder()
	sf, _ := builder.Build(reflect.TypeOf(TestStruct{}))
	ts := TestStruct{
		Name:               "dd",
		StringWithOption:   "ceph",
		StringWithLenLimit: "aaa",
		IntWithRange:       100,
		Composition: []IncludeStruct{
			IncludeStruct{
				Int8WithRange: 5,
			},
		},
	}
	ts.StringMapCompostion = make(map[string]IncludeStruct)
	ts.StringMapCompostion["a"] = IncludeStruct{
		Int8WithRange: 10,
	}

	ts.IntMapCompostion = make(map[int32]IncludeStruct)
	ts.IntMapCompostion[20] = IncludeStruct{
		Int8WithRange: 19,
	}

	ut.Assert(t, sf.Validate(ts) == nil, "")

	ts2 := ts
	ts2.StringWithOption = "oo"
	ut.Assert(t, sf.Validate(ts2) != nil, "")

	ts3 := ts
	ts3.IntWithRange = 10000
	ut.Assert(t, sf.Validate(ts3) != nil, "")

	ts4 := ts
	ss := ts4.StringMapCompostion["a"]
	ss.Int8WithRange = 22
	ts4.StringMapCompostion["a"] = ss
	ut.Assert(t, sf.Validate(ts4) != nil, "")

	ts5 := ts
	ss = ts5.IntMapCompostion[20]
	ss.Int8WithRange = -1
	ts5.IntMapCompostion[20] = ss
	ut.Assert(t, sf.Validate(ts5) != nil, "")
}
