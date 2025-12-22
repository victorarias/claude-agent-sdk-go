package subprocess

import (
	"testing"
)

func TestValidatePath_RejectsTraversal(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"clean absolute path", "/usr/local/bin", false},
		{"clean relative path", "plugins/my-plugin", false},
		{"current directory", ".", false},
		{"parent traversal", "../etc/passwd", true},
		{"hidden parent traversal", "foo/../../../etc/passwd", true},
		{"double dot in middle", "foo/bar/../baz", true},
		{"just double dot", "..", true},
		{"empty path", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePath(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
		})
	}
}

func TestValidatePath_RejectsNullBytes(t *testing.T) {
	err := ValidatePath("/path/with\x00null")
	if err == nil {
		t.Error("ValidatePath should reject paths with null bytes")
	}
}
