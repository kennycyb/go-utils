package strutil

import (
	"testing"
)

func TestIsEmpty(t *testing.T) {

	var s *string

	if !IsEmpty(s) {
		t.Errorf("IsEmpty failed")
	}

	x := ""
	s = &x

	if !IsEmpty(s) {
		t.Errorf("IsEmpty failed")
	}
}
