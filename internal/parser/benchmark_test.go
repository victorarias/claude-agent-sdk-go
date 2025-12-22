// Copyright (C) 2025 Claude Agent SDK Go Contributors
// SPDX-License-Identifier: GPL-3.0-only

package parser

import (
	"testing"
)

// Benchmark fixtures for various message types
var (
	systemMessageRaw = map[string]any{
		"type":    "system",
		"subtype": "init",
		"data": map[string]any{
			"version":    "2.0.0",
			"session_id": "bench_session_123",
		},
	}

	assistantMessageRaw = map[string]any{
		"type": "assistant",
		"message": map[string]any{
			"content": []any{
				map[string]any{"type": "text", "text": "Hello! This is a benchmark test message."},
			},
			"model": "claude-sonnet-4-5",
		},
	}

	assistantMessageWithToolUse = map[string]any{
		"type": "assistant",
		"message": map[string]any{
			"content": []any{
				map[string]any{"type": "text", "text": "Let me help you with that."},
				map[string]any{
					"type":  "tool_use",
					"id":    "tool_123",
					"name":  "bash",
					"input": map[string]any{"command": "ls -la"},
				},
			},
			"model": "claude-sonnet-4-5",
		},
	}

	resultMessageRaw = map[string]any{
		"type":           "result",
		"subtype":        "success",
		"duration_ms":    float64(1500),
		"session_id":     "bench_session_123",
		"total_cost_usd": float64(0.0025),
		"num_turns":      float64(3),
	}

	userMessageRaw = map[string]any{
		"type": "user",
		"message": map[string]any{
			"content": []any{
				map[string]any{"type": "text", "text": "What is the weather today?"},
			},
		},
	}

	streamEventRaw = map[string]any{
		"type":       "stream_event",
		"uuid":       "event_123",
		"session_id": "bench_session_123",
		"content_block": map[string]any{
			"type": "text",
			"text": "Streaming content...",
		},
	}
)

func BenchmarkParseMessage_System(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = ParseMessage(systemMessageRaw)
	}
}

func BenchmarkParseMessage_Assistant(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = ParseMessage(assistantMessageRaw)
	}
}

func BenchmarkParseMessage_AssistantWithToolUse(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = ParseMessage(assistantMessageWithToolUse)
	}
}

func BenchmarkParseMessage_Result(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = ParseMessage(resultMessageRaw)
	}
}

func BenchmarkParseMessage_User(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = ParseMessage(userMessageRaw)
	}
}

func BenchmarkParseMessage_StreamEvent(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = ParseMessage(streamEventRaw)
	}
}

// BenchmarkParseMessage_Mixed simulates real-world usage with mixed message types
func BenchmarkParseMessage_Mixed(b *testing.B) {
	messages := []map[string]any{
		systemMessageRaw,
		assistantMessageRaw,
		userMessageRaw,
		assistantMessageWithToolUse,
		streamEventRaw,
		resultMessageRaw,
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		msg := messages[i%len(messages)]
		_, _ = ParseMessage(msg)
	}
}
