// Copyright (C) 2025 Claude Agent SDK Go Contributors
// SPDX-License-Identifier: GPL-3.0-only

package subprocess

import (
	"testing"

	"github.com/victorarias/claude-agent-sdk-go/types"
)

func TestValidateOptionSemantics(t *testing.T) {
	tests := []struct {
		name      string
		opts      *types.Options
		shouldErr bool
	}{
		{
			name: "bypass without allow-dangerous flag",
			opts: &types.Options{
				PermissionMode: types.PermissionBypass,
			},
			shouldErr: true,
		},
		{
			name: "same model and fallback model",
			opts: &types.Options{
				Model:         "claude-sonnet-4-5",
				FallbackModel: "claude-sonnet-4-5",
			},
			shouldErr: true,
		},
		{
			name: "continue and resume together",
			opts: &types.Options{
				ContinueConversation: true,
				Resume:               "sess-1",
			},
			shouldErr: true,
		},
		{
			name: "resume_session_at without resume",
			opts: &types.Options{
				ResumeSessionAt: "msg-1",
			},
			shouldErr: true,
		},
		{
			name: "session_id with resume without fork_session",
			opts: &types.Options{
				SessionID: "sess-2",
				Resume:    "sess-1",
			},
			shouldErr: true,
		},
		{
			name: "valid configuration",
			opts: &types.Options{
				PermissionMode:                  types.PermissionBypass,
				AllowDangerouslySkipPermissions: true,
				Model:                           "claude-opus-4",
				FallbackModel:                   "claude-sonnet-4-5",
				Resume:                          "sess-1",
				ResumeSessionAt:                 "msg-1",
				SessionID:                       "sess-2",
				ForkSession:                     true,
			},
			shouldErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateOptionSemantics(tt.opts)
			if tt.shouldErr && err == nil {
				t.Fatal("expected validation error, got nil")
			}
			if !tt.shouldErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}
