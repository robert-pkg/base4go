package utils

import (
	"testing"
)

func TestParseAddr(t *testing.T) {
	raw := "consul://Greeter/Echo"
	scheme, service, method, err := ParseAddr(raw)
	if err != nil {
		t.Errorf("ParseAddr(%v) error = %v", raw, err)
	}
	if scheme != "consul" {
		t.Errorf("ParseAddr(%v) scheme = %v, want %v", raw, scheme, "consul")
	}
	if service != "Greeter" {
		t.Errorf("ParseAddr(%v) service = %v, want %v", raw, service, "Greeter")
	}
	if method != "Echo" {
		t.Errorf("ParseAddr(%v) method = %v, want %v", raw, method, "Echo")
	}
}
