package resourcefield

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

	ut.Equal(t, len(sf.fields), 12)
	fieldNames := []string{
		"Id",
		"Age",
		"Name",
		"StringWithOption",
		"StringWithLenLimit",
		"IntWithRange",

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
		StringWithLenLimit string `json:"stringWithLenLimit" rest:"minLen=20,maxLen=10"`
	}
	builder := NewBuilder()
	_, err := builder.Build(reflect.TypeOf(S1{}))
	ut.Assert(t, err != nil, "")

	type S2 struct {
		IntWithRange uint32 `json:"intWithRange" rest:"min=100,max=10"`
	}
	builder = NewBuilder()
	_, err = builder.Build(reflect.TypeOf(S2{}))
	ut.Assert(t, err != nil, "")
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