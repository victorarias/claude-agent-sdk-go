// Copyright (C) 2025 Claude Agent SDK Go Contributors
// SPDX-License-Identifier: GPL-3.0-only

package sdk

import (
	"testing"
	"time"
)

func TestResolveInitializeTimeout_Default(t *testing.T) {
	t.Setenv("CLAUDE_CODE_STREAM_CLOSE_TIMEOUT", "")
	if got := resolveInitializeTimeout(); got != 60*time.Second {
		t.Fatalf("expected 60s, got %v", got)
	}
}

func TestResolveInitializeTimeout_UsesEnvWhenAboveMinimum(t *testing.T) {
	t.Setenv("CLAUDE_CODE_STREAM_CLOSE_TIMEOUT", "120000")
	if got := resolveInitializeTimeout(); got != 120*time.Second {
		t.Fatalf("expected 120s, got %v", got)
	}
}

func TestResolveInitializeTimeout_ClampsToMinimum(t *testing.T) {
	t.Setenv("CLAUDE_CODE_STREAM_CLOSE_TIMEOUT", "1000")
	if got := resolveInitializeTimeout(); got != 60*time.Second {
		t.Fatalf("expected clamp to 60s, got %v", got)
	}
}

func TestResolveInitializeTimeout_InvalidValue(t *testing.T) {
	t.Setenv("CLAUDE_CODE_STREAM_CLOSE_TIMEOUT", "invalid")
	if got := resolveInitializeTimeout(); got != 60*time.Second {
		t.Fatalf("expected 60s for invalid env value, got %v", got)
	}
}
