package internal

import (
	"testing"
)

func TestGetArg(t *testing.T) {
	args := []string{"foo", "bar", "baz"}

	if v, f := GetArg(0, args...); !f || v != "foo" {
		t.Errorf("invalid return value")
	}

	if v, f := GetArg(1, args...); !f || v != "bar" {
		t.Errorf("invalid return value")
	}

	if v, f := GetArg(2, args...); !f || v != "baz" {
		t.Errorf("invalid return value")
	}

	if v, f := GetArg(3, args...); f || v != "" {
		t.Errorf("invalid return value")
	}
}
