package resourcefield

import (
	"reflect"
	"strings"
	"testing"

	ut "github.com/zdnscloud/cement/unittest"
)

func TestBuildValidatorWithValidTag(t *testing.T) {
	sv := reflect.ValueOf(TestStruct{
		StringWithOption:   "lvm",
		StringWithLenLimit: "good",
		IntWithRange:       10,
	})

	sv2 := reflect.ValueOf(TestStruct{
		StringWithOption:   "lvms",
		StringWithLenLimit: "g",
		IntWithRange:       10000,
	})
	st := sv.Type()
	vc := 0
	for i := 0; i < st.NumField(); i++ {
		f := st.Field(i)
		tags := strings.Split(f.Tag.Get("rest"), ",")
		if len(tags) > 0 {
			validator, err := buildValidator(f.Type.Kind(), tags)
			ut.Assert(t, err == nil, "get err %v", err)
			if validator != nil {
				vc += 1
				ut.Assert(t, validator.Validate(sv.Field(i).Interface()) == nil, "")
				ut.Assert(t, validator.Validate(sv2.Field(i).Interface()) != nil, "")
			}
		}
	}
	ut.Equal(t, vc, 3)
}

func TestBuildValidatorWithInValidTag(t *testing.T) {
	type testStruct struct {
		IntWithOption      int    `rest:"required=true,options=lvm|ceph"`
		IntWithLenLimit    int    `rest:"minLen=10,maxLen=11"`
		StringWithLenLimit string `rest:"minLen=12,maxLen=12"`
		ShortOfMax         uint32 `rest:"min=1"`
		ShortOfMin         int8   `rest:"max=1"`
	}

	sv := reflect.ValueOf(testStruct{})
	st := sv.Type()
	for i := 0; i < st.NumField(); i++ {
		f := st.Field(i)
		tags := strings.Split(f.Tag.Get("rest"), ",")
		if len(tags) > 0 {
			validator, _ := buildValidator(f.Type.Kind(), tags)
			ut.Assert(t, validator == nil, "")
		}
	}
}
