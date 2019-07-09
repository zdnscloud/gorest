package util

import (
	"io"
	"testing"
)

func TestNilCheck(t *testing.T) {
	if IsValueNil(nil) == false {
		t.Fatal("nil isn't nil")
	}

	var m map[string]string
	if IsValueNil(m) == false {
		t.Fatal("empty map is nil")
	}

	type S struct{}
	var sp *S
	var ss S
	if IsValueNil(sp) == false {
		t.Fatal("empty struct pointer is nil")
	}
	if IsValueNil(ss) == true {
		t.Fatal("empty struct isn't nil")
	}

	var p io.Reader
	if IsValueNil(p) == false {
		t.Fatal("empty interface is nil")
	}
}
