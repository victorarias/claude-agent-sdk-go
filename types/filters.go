package types

// FilterTextBlocks filters a slice of ContentBlocks and returns only TextBlocks.
// Returns an empty slice if no TextBlocks are found.
func FilterTextBlocks(blocks []ContentBlock) []*TextBlock {
	result := make([]*TextBlock, 0)
	for _, block := range blocks {
		if textBlock, ok := block.(*TextBlock); ok {
			result = append(result, textBlock)
		}
	}
	return result
}

// FilterToolUseBlocks filters a slice of ContentBlocks and returns only ToolUseBlocks.
// Returns an empty slice if no ToolUseBlocks are found.
func FilterToolUseBlocks(blocks []ContentBlock) []*ToolUseBlock {
	result := make([]*ToolUseBlock, 0)
	for _, block := range blocks {
		if toolUseBlock, ok := block.(*ToolUseBlock); ok {
			result = append(result, toolUseBlock)
		}
	}
	return result
}

// FilterToolResultBlocks filters a slice of ContentBlocks and returns only ToolResultBlocks.
// Returns an empty slice if no ToolResultBlocks are found.
func FilterToolResultBlocks(blocks []ContentBlock) []*ToolResultBlock {
	result := make([]*ToolResultBlock, 0)
	for _, block := range blocks {
		if toolResultBlock, ok := block.(*ToolResultBlock); ok {
			result = append(result, toolResultBlock)
		}
	}
	return result
}

// FilterThinkingBlocks filters a slice of ContentBlocks and returns only ThinkingBlocks.
// Returns an empty slice if no ThinkingBlocks are found.
func FilterThinkingBlocks(blocks []ContentBlock) []*ThinkingBlock {
	result := make([]*ThinkingBlock, 0)
	for _, block := range blocks {
		if thinkingBlock, ok := block.(*ThinkingBlock); ok {
			result = append(result, thinkingBlock)
		}
	}
	return result
}
