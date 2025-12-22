// Copyright (C) 2025 Claude Agent SDK Go Contributors
// SPDX-License-Identifier: GPL-3.0-only

package subprocess

import (
	"testing"

	"github.com/victorarias/claude-agent-sdk-go/types"
)

func TestBuildCommand_ForkSession(t *testing.T) {
	tests := []struct {
		name        string
		forkSession bool
		shouldHave  bool
	}{
		{
			name:        "fork session enabled",
			forkSession: true,
			shouldHave:  true,
		},
		{
			name:        "fork session disabled",
			forkSession: false,
			shouldHave:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := types.DefaultOptions()
			opts.ForkSession = tt.forkSession

			cmd := buildCommand("/usr/bin/claude", "test prompt", opts, false)

			hasForkSession := false
			for _, arg := range cmd {
				if arg == "--fork-session" {
					hasForkSession = true
					break
				}
			}

			if hasForkSession != tt.shouldHave {
				if tt.shouldHave {
					t.Errorf("expected --fork-session flag to be present but it was not")
				} else {
					t.Errorf("expected --fork-session flag to be absent but it was present")
				}
			}
		})
	}
}

func TestWithForkSession(t *testing.T) {
	opts := types.DefaultOptions()

	// Apply the WithForkSession option
	types.ApplyOptions(opts, types.WithForkSession())

	if !opts.ForkSession {
		t.Errorf("WithForkSession() should set ForkSession to true, got false")
	}
}
