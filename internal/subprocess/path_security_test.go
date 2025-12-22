package subprocess

import (
	"testing"

	"github.com/victorarias/claude-agent-sdk-go/types"
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

func TestValidatePathOptions_RejectsTraversalInAddDirs(t *testing.T) {
	opts := types.DefaultOptions()
	opts.AddDirs = []string{"../../../etc"}

	err := ValidatePathOptions(opts)
	if err == nil {
		t.Error("ValidatePathOptions should reject AddDirs with path traversal")
	}
}

func TestValidatePathOptions_RejectsTraversalInPluginPath(t *testing.T) {
	opts := types.DefaultOptions()
	opts.Plugins = []types.PluginConfig{
		{Type: "local", Path: "../../../etc/passwd"},
	}

	err := ValidatePathOptions(opts)
	if err == nil {
		t.Error("ValidatePathOptions should reject plugin paths with path traversal")
	}
}

func TestValidatePathOptions_AcceptsValidPaths(t *testing.T) {
	opts := types.DefaultOptions()
	opts.AddDirs = []string{"/usr/local/share", "relative/path"}
	opts.Plugins = []types.PluginConfig{
		{Type: "local", Path: "/usr/local/plugins/my-plugin"},
	}

	err := ValidatePathOptions(opts)
	if err != nil {
		t.Errorf("ValidatePathOptions should accept valid paths, got error: %v", err)
	}
}
