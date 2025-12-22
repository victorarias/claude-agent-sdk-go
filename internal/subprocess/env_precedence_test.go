// Copyright (C) 2025 Claude Agent SDK Go Contributors
// SPDX-License-Identifier: GPL-3.0-only

package subprocess

import (
	"os"
	"strings"
	"testing"

	"github.com/victorarias/claude-agent-sdk-go/types"
)

// TestEnvironmentPrecedence verifies that environment variables follow the correct precedence:
// 1. System environment (os.Environ)
// 2. User-provided env overrides (from Options.Env)
// 3. SDK-required env (TERM=dumb, NO_COLOR=1) - these CANNOT be overridden
func TestEnvironmentPrecedence(t *testing.T) {
	tests := []struct {
		name     string
		systemEnv map[string]string
		userEnv   map[string]string
		want      map[string]string
	}{
		{
			name: "SDK env vars set after user env",
			systemEnv: map[string]string{
				"PATH": "/usr/bin",
			},
			userEnv: map[string]string{
				"CUSTOM": "value",
			},
			want: map[string]string{
				"PATH":   "/usr/bin",
				"CUSTOM": "value",
				"TERM":   "dumb",
				"NO_COLOR": "1",
				"CLAUDE_CODE_ENTRYPOINT": "sdk-go",
			},
		},
		{
			name: "user env cannot override TERM",
			systemEnv: map[string]string{
				"PATH": "/usr/bin",
			},
			userEnv: map[string]string{
				"TERM": "xterm-256color", // User tries to set TERM
			},
			want: map[string]string{
				"PATH":   "/usr/bin",
				"TERM":   "dumb", // SDK overrides to dumb
				"NO_COLOR": "1",
			},
		},
		{
			name: "user env cannot override NO_COLOR",
			systemEnv: map[string]string{
				"PATH": "/usr/bin",
			},
			userEnv: map[string]string{
				"NO_COLOR": "0", // User tries to enable color
			},
			want: map[string]string{
				"PATH":   "/usr/bin",
				"TERM":   "dumb",
				"NO_COLOR": "1", // SDK forces NO_COLOR=1
			},
		},
		{
			name: "user env overrides system env",
			systemEnv: map[string]string{
				"PATH":   "/usr/bin",
				"CUSTOM": "system-value",
			},
			userEnv: map[string]string{
				"CUSTOM": "user-value", // User overrides system
			},
			want: map[string]string{
				"PATH":   "/usr/bin",
				"CUSTOM": "user-value", // User value takes precedence over system
				"TERM":   "dumb",
				"NO_COLOR": "1",
			},
		},
		{
			name: "SDK internal vars cannot be overridden",
			systemEnv: map[string]string{
				"PATH": "/usr/bin",
			},
			userEnv: map[string]string{
				"CLAUDE_CODE_ENTRYPOINT": "malicious", // User tries to override
			},
			want: map[string]string{
				"PATH": "/usr/bin",
				"CLAUDE_CODE_ENTRYPOINT": "sdk-go", // SDK maintains control
				"TERM":   "dumb",
				"NO_COLOR": "1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup system environment
			for k, v := range tt.systemEnv {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			// Create options with user env
			opts := types.DefaultOptions()
			opts.Env = tt.userEnv

			// Build environment
			env := buildEnvironment(opts)

			// Parse env slice into map for easier verification
			envMap := make(map[string]string)
			for _, e := range env {
				parts := strings.SplitN(e, "=", 2)
				if len(parts) == 2 {
					envMap[parts[0]] = parts[1]
				}
			}

			// Verify expected values
			for key, expectedValue := range tt.want {
				actualValue, exists := envMap[key]
				if !exists {
					t.Errorf("expected env var %q to exist, but it was not found", key)
					continue
				}
				if actualValue != expectedValue {
					t.Errorf("env var %q: expected %q, got %q", key, expectedValue, actualValue)
				}
			}
		})
	}
}

// TestSDKEnvironmentVariablesAlwaysSet verifies that SDK-required env vars are always set
func TestSDKEnvironmentVariablesAlwaysSet(t *testing.T) {
	opts := types.DefaultOptions()
	env := buildEnvironment(opts)

	// Parse env into map
	envMap := make(map[string]string)
	for _, e := range env {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}

	// Verify SDK-required vars are always present
	requiredVars := map[string]string{
		"TERM":                     "dumb",
		"NO_COLOR":                 "1",
		"CLAUDE_CODE_ENTRYPOINT":   "sdk-go",
		"CLAUDE_AGENT_SDK_VERSION": types.Version,
	}

	for key, expectedValue := range requiredVars {
		actualValue, exists := envMap[key]
		if !exists {
			t.Errorf("SDK-required env var %q is missing", key)
			continue
		}
		if actualValue != expectedValue {
			t.Errorf("SDK-required env var %q: expected %q, got %q", key, expectedValue, actualValue)
		}
	}
}

// TestFeatureFlagEnvironment verifies feature flags are set correctly
func TestFeatureFlagEnvironment(t *testing.T) {
	opts := types.DefaultOptions()
	opts.EnableFileCheckpointing = true

	env := buildEnvironment(opts)

	// Parse env into map
	envMap := make(map[string]string)
	for _, e := range env {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}

	// Verify feature flag is set
	if val, exists := envMap["CLAUDE_CODE_ENABLE_SDK_FILE_CHECKPOINTING"]; !exists {
		t.Error("expected CLAUDE_CODE_ENABLE_SDK_FILE_CHECKPOINTING to be set")
	} else if val != "true" {
		t.Errorf("expected CLAUDE_CODE_ENABLE_SDK_FILE_CHECKPOINTING=true, got %q", val)
	}

	// Verify it cannot be overridden by user
	opts.Env = map[string]string{
		"CLAUDE_CODE_ENABLE_SDK_FILE_CHECKPOINTING": "false",
	}
	env = buildEnvironment(opts)

	envMap = make(map[string]string)
	for _, e := range env {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}

	// SDK should maintain control
	if val := envMap["CLAUDE_CODE_ENABLE_SDK_FILE_CHECKPOINTING"]; val != "true" {
		t.Errorf("user should not be able to override SDK feature flag, expected 'true' got %q", val)
	}
}
