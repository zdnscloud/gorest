package resourcefield

import (
	"reflect"
	"strings"
	"testing"

	ut "github.com/zdnscloud/cement/unittest"
)

func TestBuildValidator(t *testing.T) {
	type testStruct struct {
		DomainName            string     `json:"domainName" rest:"isDomain=true"`
		StringWithOption      MyOption   `json:"stringWithOption,omitempty" rest:"required=true,options=lvm|ceph"`
		StringWithLenLimit    string     `json:"stringWithLenLimit" rest:"minLen=2,maxLen=10"`
		IntWithRange          uint32     `json:"intWithRange" rest:"min=1,max=1000"`
		StringSliceWithDomain []string   `json:"stringSliceWithDomain,omitempty" rest:"required=true,isDomain=true"`
		StringSliceWithOption []MyOption `json:"stringSliceWithOption,omitempty" rest:"required=true,options=lvm|ceph"`
	}

	st := reflect.TypeOf(testStruct{})
	for i := 0; i < st.NumField(); i++ {
		f := st.Field(i)
		tags := strings.Split(f.Tag.Get("rest"), ",")
		ut.Assert(t, len(tags) > 0, "")
		validator, err := buildValidator(f.Type, tags)
		ut.Assert(t, err == nil && validator != nil, "")
	}

	type testStruct2 struct {
		IntWithOption      int    `rest:"required=true,options=lvm|ceph"`
		IntWithLenLimit    int    `rest:"minLen=10,maxLen=11"`
		IntWithDomainCheck int    `rest:"isDomain=true"`
		IntRangeShortOfMax uint32 `rest:"min=1"`
		IntRangeShortOfMin int8   `rest:"max=1"`
		IntRangeInvalid    int8   `rest:"min=2,max=1"`

		StringWithInvalidLenLimit       string `rest:"minLen=12,maxLen=12"`
		StringWithPartialLenLimit       string `rest:"minLen=12"`
		StringWithBothLenLimitAndDoamin string `rest:"minLen=1,maxLen=10,isDomain=true"`
		StringWithBothOptionAndDoamin   string `rest:"options=xx|bb,isDomain=true"`
		StringWithBothOptionAndLenLimit string `rest:"options=xx|bb,minLen=1,maxLen=10"`
	}

	st = reflect.TypeOf(testStruct2{})
	for i := 0; i < st.NumField(); i++ {
		f := st.Field(i)
		tags := strings.Split(f.Tag.Get("rest"), ",")
		validator, _ := buildValidator(f.Type, tags)
		ut.Assert(t, validator == nil, "%d:%v", i, f)
	}
}

type testCase struct {
	value    interface{}
	isValide bool
}

func TestIntegerRangeValidator(t *testing.T) {
	validator, err := buildValidator(reflect.TypeOf(uint(10)), []string{"min=1", "max=10"})
	ut.Assert(t, err == nil && validator != nil, "")
	testValidator(t, validator, []testCase{
		{1, true},
		{10, false},
		{11, false},
		{[]int{1, 2, 9}, true},
		{[]int{1, 2, 10}, false},
		{[]int{10, 11}, false},
	})
}

func TestStringLenValidator(t *testing.T) {
	validator, err := buildValidator(reflect.TypeOf("xxx"), []string{"minLen=1", "maxLen=3"})
	ut.Assert(t, err == nil && validator != nil, "")
	testValidator(t, validator, []testCase{
		{"a", true},
		{"abc", false},
		{"", false},
		{[]string{"a", "ab", "b"}, true},
		{[]string{"a", "abc", "b"}, false},
		{[]string{"", "abc"}, false},
	})
}

func TestOptionValidator(t *testing.T) {
	validator, err := buildValidator(reflect.TypeOf("xxx"), []string{"options=aa|bb"})
	ut.Assert(t, err == nil && validator != nil, "")
	testValidator(t, validator, []testCase{
		{"aa", true},
		{"bb", true},
		{"Aa", false},
		{"Aaa", false},
		{[]string{"aa", "bb", "aa"}, true},
		{[]string{"xa", "bb"}, false},
		{[]string{"", "ac"}, false},
	})
}

func TestDomainValidator(t *testing.T) {
	validator, err := buildValidator(reflect.TypeOf("xxx"), []string{"isDomain=true"})
	ut.Assert(t, err == nil && validator != nil, "")
	testValidator(t, validator, []testCase{
		{"aa", true},
		{"11-bb", true},
		{"11_bb", false},
		{"Aaa", false},
		{"-aa", false},
		{"11?aa", false},
		{[]string{"aa", "11bb", "11.aa"}, true},
		{[]string{"adsfasdf-xa", "11111bb11111"}, true},
		{[]string{"adsfasdfasdfA111", "adsfadf?xx"}, false},
	})
}

func testValidator(t *testing.T, validator Validator, cases []testCase) {
	for i := 0; i < len(cases); i++ {
		err := validator.Validate(cases[i].value)
		if cases[i].isValide {
			ut.Assert(t, err == nil, "")
		} else {
			ut.Assert(t, err != nil, "")
		}
	}
}
