package types

import (
	"testing"
)

func TestFilterTextBlocks(t *testing.T) {
	tests := []struct {
		name     string
		blocks   []ContentBlock
		expected []*TextBlock
	}{
		{
			name:     "empty slice returns empty slice",
			blocks:   []ContentBlock{},
			expected: []*TextBlock{},
		},
		{
			name: "single text block",
			blocks: []ContentBlock{
				&TextBlock{TextContent: "hello"},
			},
			expected: []*TextBlock{
				{TextContent: "hello"},
			},
		},
		{
			name: "mixed blocks returns only text",
			blocks: []ContentBlock{
				&TextBlock{TextContent: "first"},
				&ToolUseBlock{ID: "tool1", Name: "bash"},
				&TextBlock{TextContent: "second"},
				&ThinkingBlock{ThinkingContent: "thinking"},
				&TextBlock{TextContent: "third"},
			},
			expected: []*TextBlock{
				{TextContent: "first"},
				{TextContent: "second"},
				{TextContent: "third"},
			},
		},
		{
			name: "no text blocks returns empty",
			blocks: []ContentBlock{
				&ToolUseBlock{ID: "tool1", Name: "bash"},
				&ThinkingBlock{ThinkingContent: "thinking"},
			},
			expected: []*TextBlock{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterTextBlocks(tt.blocks)
			if len(result) != len(tt.expected) {
				t.Errorf("FilterTextBlocks() returned %d blocks, expected %d", len(result), len(tt.expected))
				return
			}
			for i, block := range result {
				if block.TextContent != tt.expected[i].TextContent {
					t.Errorf("FilterTextBlocks()[%d].TextContent = %q, expected %q", i, block.TextContent, tt.expected[i].TextContent)
				}
			}
		})
	}
}

func TestFilterToolUseBlocks(t *testing.T) {
	tests := []struct {
		name     string
		blocks   []ContentBlock
		expected []*ToolUseBlock
	}{
		{
			name:     "empty slice returns empty slice",
			blocks:   []ContentBlock{},
			expected: []*ToolUseBlock{},
		},
		{
			name: "single tool use block",
			blocks: []ContentBlock{
				&ToolUseBlock{ID: "tool1", Name: "bash"},
			},
			expected: []*ToolUseBlock{
				{ID: "tool1", Name: "bash"},
			},
		},
		{
			name: "mixed blocks returns only tool use",
			blocks: []ContentBlock{
				&TextBlock{TextContent: "first"},
				&ToolUseBlock{ID: "tool1", Name: "bash"},
				&ThinkingBlock{ThinkingContent: "thinking"},
				&ToolUseBlock{ID: "tool2", Name: "read"},
				&TextBlock{TextContent: "second"},
			},
			expected: []*ToolUseBlock{
				{ID: "tool1", Name: "bash"},
				{ID: "tool2", Name: "read"},
			},
		},
		{
			name: "no tool use blocks returns empty",
			blocks: []ContentBlock{
				&TextBlock{TextContent: "hello"},
				&ThinkingBlock{ThinkingContent: "thinking"},
			},
			expected: []*ToolUseBlock{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterToolUseBlocks(tt.blocks)
			if len(result) != len(tt.expected) {
				t.Errorf("FilterToolUseBlocks() returned %d blocks, expected %d", len(result), len(tt.expected))
				return
			}
			for i, block := range result {
				if block.ID != tt.expected[i].ID || block.Name != tt.expected[i].Name {
					t.Errorf("FilterToolUseBlocks()[%d] = {ID: %q, Name: %q}, expected {ID: %q, Name: %q}",
						i, block.ID, block.Name, tt.expected[i].ID, tt.expected[i].Name)
				}
			}
		})
	}
}

func TestFilterToolResultBlocks(t *testing.T) {
	tests := []struct {
		name     string
		blocks   []ContentBlock
		expected []*ToolResultBlock
	}{
		{
			name:     "empty slice returns empty slice",
			blocks:   []ContentBlock{},
			expected: []*ToolResultBlock{},
		},
		{
			name: "single tool result block",
			blocks: []ContentBlock{
				&ToolResultBlock{ToolUseID: "tool1", ResultContent: "output"},
			},
			expected: []*ToolResultBlock{
				{ToolUseID: "tool1", ResultContent: "output"},
			},
		},
		{
			name: "mixed blocks returns only tool result",
			blocks: []ContentBlock{
				&TextBlock{TextContent: "first"},
				&ToolResultBlock{ToolUseID: "tool1", ResultContent: "output1"},
				&ToolUseBlock{ID: "tool2", Name: "bash"},
				&ToolResultBlock{ToolUseID: "tool3", ResultContent: "output2", IsError: true},
			},
			expected: []*ToolResultBlock{
				{ToolUseID: "tool1", ResultContent: "output1"},
				{ToolUseID: "tool3", ResultContent: "output2", IsError: true},
			},
		},
		{
			name: "no tool result blocks returns empty",
			blocks: []ContentBlock{
				&TextBlock{TextContent: "hello"},
				&ToolUseBlock{ID: "tool1", Name: "bash"},
			},
			expected: []*ToolResultBlock{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterToolResultBlocks(tt.blocks)
			if len(result) != len(tt.expected) {
				t.Errorf("FilterToolResultBlocks() returned %d blocks, expected %d", len(result), len(tt.expected))
				return
			}
			for i, block := range result {
				if block.ToolUseID != tt.expected[i].ToolUseID ||
					block.ResultContent != tt.expected[i].ResultContent ||
					block.IsError != tt.expected[i].IsError {
					t.Errorf("FilterToolResultBlocks()[%d] = {ToolUseID: %q, ResultContent: %q, IsError: %v}, expected {ToolUseID: %q, ResultContent: %q, IsError: %v}",
						i, block.ToolUseID, block.ResultContent, block.IsError,
						tt.expected[i].ToolUseID, tt.expected[i].ResultContent, tt.expected[i].IsError)
				}
			}
		})
	}
}

func TestFilterThinkingBlocks(t *testing.T) {
	tests := []struct {
		name     string
		blocks   []ContentBlock
		expected []*ThinkingBlock
	}{
		{
			name:     "empty slice returns empty slice",
			blocks:   []ContentBlock{},
			expected: []*ThinkingBlock{},
		},
		{
			name: "single thinking block",
			blocks: []ContentBlock{
				&ThinkingBlock{ThinkingContent: "hmm", Signature: "sig1"},
			},
			expected: []*ThinkingBlock{
				{ThinkingContent: "hmm", Signature: "sig1"},
			},
		},
		{
			name: "mixed blocks returns only thinking",
			blocks: []ContentBlock{
				&TextBlock{TextContent: "first"},
				&ThinkingBlock{ThinkingContent: "thinking1", Signature: "sig1"},
				&ToolUseBlock{ID: "tool1", Name: "bash"},
				&ThinkingBlock{ThinkingContent: "thinking2", Signature: "sig2"},
			},
			expected: []*ThinkingBlock{
				{ThinkingContent: "thinking1", Signature: "sig1"},
				{ThinkingContent: "thinking2", Signature: "sig2"},
			},
		},
		{
			name: "no thinking blocks returns empty",
			blocks: []ContentBlock{
				&TextBlock{TextContent: "hello"},
				&ToolUseBlock{ID: "tool1", Name: "bash"},
			},
			expected: []*ThinkingBlock{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterThinkingBlocks(tt.blocks)
			if len(result) != len(tt.expected) {
				t.Errorf("FilterThinkingBlocks() returned %d blocks, expected %d", len(result), len(tt.expected))
				return
			}
			for i, block := range result {
				if block.ThinkingContent != tt.expected[i].ThinkingContent ||
					block.Signature != tt.expected[i].Signature {
					t.Errorf("FilterThinkingBlocks()[%d] = {ThinkingContent: %q, Signature: %q}, expected {ThinkingContent: %q, Signature: %q}",
						i, block.ThinkingContent, block.Signature,
						tt.expected[i].ThinkingContent, tt.expected[i].Signature)
				}
			}
		})
	}
}
