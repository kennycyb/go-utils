package strutil

import (
	"testing"
)

func TestSnakeCase(t *testing.T) {
	if ToSnakeCase("HelloWorld") != "hello_world" {
		t.Errorf("ToSnakeCase failed")
	}
}
