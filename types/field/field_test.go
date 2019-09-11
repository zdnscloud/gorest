package field

import (
	"encoding/json"
	ut "github.com/zdnscloud/cement/unittest"
	"reflect"
	"testing"
)

func TestFieldBuild(t *testing.T) {
	builder := NewBuilder()
	sf, err := builder.Build(reflect.TypeOf(TestStruct{}))
	ut.Assert(t, err == nil, "")

	ut.Equal(t, len(sf.fields), 15)
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

		"SliceComposition",
		"StringMapCompostion",
		"IntMapCompostion",

		"PtrComposition",
		"SlicePtrComposition",
		"IntPtrMapCompostion",
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
		SliceComposition: []IncludeStruct{
			IncludeStruct{
				Int8WithRange: 7,
			},
		},
		StringMapCompostion: map[string]IncludeStruct{
			"a": IncludeStruct{
				Int8WithRange: 6,
			},
		},
		IntMapCompostion: map[int32]IncludeStruct{
			10: IncludeStruct{
				Int8WithRange: 6,
			},
		},
		PtrComposition: &IncludeStruct{
			Int8WithRange: 7,
		},
		SlicePtrComposition: []*IncludeStruct{
			&IncludeStruct{
				Int8WithRange: 5,
			},
		},
		IntPtrMapCompostion: map[uint8]*IncludeStruct{
			8: &IncludeStruct{
				Int8WithRange: 5,
			},
		},
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
	ut.Equal(t, ts.IntMapCompostion[10].Uint16WithDefault, uint16(11))
	ut.Equal(t, ts.StringMapCompostion["a"].Uint16WithDefault, uint16(11))
	ut.Equal(t, ts.SliceComposition[0].Uint16WithDefault, uint16(11))
	ut.Equal(t, ts.SlicePtrComposition[0].Uint16WithDefault, uint16(11))
	ut.Equal(t, ts.PtrComposition.Uint16WithDefault, uint16(11))
	ut.Equal(t, ts.IntPtrMapCompostion[8].Uint16WithDefault, uint16(11))
}

func TestCheckRequired(t *testing.T) {
	builder := NewBuilder()
	sf, _ := builder.Build(reflect.TypeOf(TestStruct{}))
	ts := TestStruct{
		Name:               "dd",
		StringWithOption:   "ceph",
		StringWithLenLimit: "aaa",
		IntWithRange:       100,
		SliceComposition: []IncludeStruct{
			IncludeStruct{
				Int8WithRange: 5,
			},
		},
		StringMapCompostion: map[string]IncludeStruct{
			"a": IncludeStruct{
				Int8WithRange: 6,
			},
		},
		IntMapCompostion: map[int32]IncludeStruct{
			10: IncludeStruct{
				Int8WithRange: 6,
			},
		},
		PtrComposition: &IncludeStruct{
			Int8WithRange: 7,
		},
		SlicePtrComposition: []*IncludeStruct{
			&IncludeStruct{
				Int8WithRange: 5,
			},
		},
		IntPtrMapCompostion: map[uint8]*IncludeStruct{
			8: &IncludeStruct{
				Int8WithRange: 5,
			},
		},
	}

	raw := make(map[string]interface{})
	rawByte, _ := json.Marshal(ts)
	json.Unmarshal(rawByte, &raw)
	ut.Assert(t, sf.CheckRequired(raw) == nil, "")

	for _, name := range []string{"name", "stringWithOption", "stringMapComposition", "intMapComposition", "ptrComposition", "slicePtrComposition", "intPtrMapComposition"} {
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
		SliceComposition: []IncludeStruct{
			IncludeStruct{
				Int8WithRange: 5,
			},
		},
		StringMapCompostion: map[string]IncludeStruct{
			"a": IncludeStruct{
				Int8WithRange: 6,
			},
		},
		IntMapCompostion: map[int32]IncludeStruct{
			10: IncludeStruct{
				Int8WithRange: 6,
			},
		},
		PtrComposition: &IncludeStruct{
			Int8WithRange: 7,
		},
		SlicePtrComposition: []*IncludeStruct{
			nil,
			&IncludeStruct{
				Int8WithRange: 5,
			},
		},
		IntPtrMapCompostion: map[uint8]*IncludeStruct{
			8: &IncludeStruct{
				Int8WithRange: 5,
			},
		},
	}

	ut.Assert(t, sf.Validate(ts) == nil, "")

	rawByte, _ := json.Marshal(ts)

	ts.StringWithOption = "oo"
	ut.Assert(t, sf.Validate(ts) != nil, "")

	ts = TestStruct{}
	json.Unmarshal(rawByte, &ts)
	ts.IntWithRange = 10000
	ut.Assert(t, sf.Validate(ts) != nil, "")

	ts = TestStruct{}
	json.Unmarshal(rawByte, &ts)
	ss := ts.StringMapCompostion["a"]
	ss.Int8WithRange = 22
	ts.StringMapCompostion["a"] = ss
	ut.Assert(t, sf.Validate(ts) != nil, "")

	ts = TestStruct{}
	json.Unmarshal(rawByte, &ts)
	ss = ts.IntMapCompostion[20]
	ss.Int8WithRange = -1
	ts.IntMapCompostion[20] = ss
	ut.Assert(t, sf.Validate(ts) != nil, "")

	ts = TestStruct{}
	json.Unmarshal(rawByte, &ts)
	ts.PtrComposition.Int8WithRange = 22
	ut.Assert(t, sf.Validate(ts) != nil, "")

	ts = TestStruct{}
	json.Unmarshal(rawByte, &ts)
	ts.SliceComposition[0].Int8WithRange = 22
	ut.Assert(t, sf.Validate(ts) != nil, "")

	ts = TestStruct{}
	json.Unmarshal(rawByte, &ts)
	ts.IntPtrMapCompostion[8].Int8WithRange = 22
	ut.Assert(t, sf.Validate(ts) != nil, "")
}
