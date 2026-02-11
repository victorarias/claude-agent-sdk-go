// Copyright (C) 2025 Claude Agent SDK Go Contributors
// SPDX-License-Identifier: GPL-3.0-only

package types

import "testing"

func TestHookEventsParityList(t *testing.T) {
	if len(HookEvents) != 15 {
		t.Fatalf("expected 15 hook events, got %d", len(HookEvents))
	}
	if HookEvents[0] != HookPreToolUse {
		t.Fatalf("expected first hook event PreToolUse, got %s", HookEvents[0])
	}
	if HookEvents[len(HookEvents)-1] != HookTaskCompleted {
		t.Fatalf("expected last hook event TaskCompleted, got %s", HookEvents[len(HookEvents)-1])
	}
}

func TestExitReasonsParityList(t *testing.T) {
	if len(ExitReasons) != 5 {
		t.Fatalf("expected 5 exit reasons, got %d", len(ExitReasons))
	}
	want := map[ExitReason]bool{
		ExitReasonClear:                     true,
		ExitReasonLogout:                    true,
		ExitReasonPromptInputExit:           true,
		ExitReasonOther:                     true,
		ExitReasonBypassPermissionsDisabled: true,
	}
	for _, reason := range ExitReasons {
		if !want[reason] {
			t.Fatalf("unexpected exit reason in list: %s", reason)
		}
	}
}
