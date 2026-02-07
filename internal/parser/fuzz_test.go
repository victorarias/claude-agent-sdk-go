// Copyright (C) 2025 Claude Agent SDK Go Contributors
// SPDX-License-Identifier: GPL-3.0-only

package parser

import (
	"encoding/json"
	"testing"
)

func FuzzParseMessage_NoPanic(f *testing.F) {
	seeds := [][]byte{
		[]byte(`{"type":"assistant","message":{"content":[{"type":"text","text":"hello"}],"model":"claude-sonnet-4-5"}}`),
		[]byte(`{"type":"user","message":{"role":"user","content":"hi"}}`),
		[]byte(`{"type":"system","subtype":"status","status":"compacting"}`),
		[]byte(`{"type":"result","subtype":"success","duration_ms":1,"duration_api_ms":1,"is_error":false,"num_turns":1,"session_id":"s"}`),
		[]byte(`{"type":"stream_event","event_type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"a"}}`),
		[]byte(`{"type":"assistant","message":{"content":[{"type":"tool_result","tool_use_id":"t1","content":[{"type":"text","text":"a"}]}]}}`),
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		var v any
		if err := json.Unmarshal(data, &v); err != nil {
			return
		}
		raw, ok := v.(map[string]any)
		if !ok {
			return
		}

		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("ParseMessage panicked: %v; input=%q", r, data)
			}
		}()

		msg, err := ParseMessage(raw)
		if err == nil && msg == nil {
			t.Fatalf("ParseMessage returned nil message without error for input=%q", data)
		}
	})
}

func FuzzParseContentBlock_NoPanic(f *testing.F) {
	seeds := [][]byte{
		[]byte(`{"type":"text","text":"hello"}`),
		[]byte(`{"type":"thinking","thinking":"hmm","signature":"sig"}`),
		[]byte(`{"type":"tool_use","id":"tool_1","name":"Bash","input":{"command":"ls"}}`),
		[]byte(`{"type":"tool_result","tool_use_id":"tool_1","content":"ok"}`),
		[]byte(`{"type":"tool_result","tool_use_id":"tool_1","content":[{"type":"text","text":"ok"}]}`),
		[]byte(`{"type":"unknown"}`),
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		var v any
		if err := json.Unmarshal(data, &v); err != nil {
			return
		}
		raw, ok := v.(map[string]any)
		if !ok {
			return
		}

		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("parseContentBlock panicked: %v; input=%q", r, data)
			}
		}()

		block, err := parseContentBlock(raw)
		if err == nil && block == nil {
			t.Fatalf("parseContentBlock returned nil block without error for input=%q", data)
		}
	})
}
