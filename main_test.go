package main

import (
	"os"
	"strings"
	"testing"
)

func TestGetBinName(t *testing.T) {
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	tests := []struct {
		arg0     string
		expected string
	}{
		{"omacorn", "omacorn"},
		{"/usr/local/bin/omacorn", "omacorn"},
		{"./dpipe", "dpipe"},
		{"dpipe", "dpipe"},
		{"", "omacorn"},
		{".", "omacorn"},
	}

	for _, tc := range tests {
		os.Args = []string{tc.arg0}
		got := getBinName()
		if got != tc.expected {
			t.Errorf("for arg0 %q, expected %q, got %q", tc.arg0, tc.expected, got)
		}
	}
}

func TestVersionConstant(t *testing.T) {
	if version == "" {
		t.Errorf("version constant cannot be empty")
	}
	if !strings.HasPrefix(version, "0.") {
		t.Errorf("unexpected version format: %s", version)
	}
}
