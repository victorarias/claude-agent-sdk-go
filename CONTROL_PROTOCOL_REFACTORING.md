# Control Protocol Typed Struct Refactoring

## Summary

Refactored the control protocol implementation in `sdk/query.go` to use typed structs instead of `map[string]any` for type safety and better code maintainability.

## Changes Made

### 1. New Typed Struct (`types/messages.go`)

Added `SDKControlMcpToolCallRequest` struct for MCP tool call requests:

```go
type SDKControlMcpToolCallRequest struct {
    Subtype    string         `json:"subtype"`
    ServerName string         `json:"server_name"`
    ToolName   string         `json:"tool_name"`
    Input      map[string]any `json:"input,omitempty"`
}
```

Updated `ParseSDKControlRequest` to handle the new `mcp_tool_call` case.

### 2. Refactored Request Handling (`sdk/query.go`)

#### `handleControlRequest`
- Now uses `types.ParseSDKControlRequest()` to parse incoming control requests into typed structs
- Uses type switch to dispatch to appropriate typed handlers
- Provides better error messages when parsing fails

#### New Typed Handlers
Created typed handler functions that accept typed structs instead of `map[string]any`:

- `handleHookCallbackTyped(*types.SDKHookCallbackRequest)` - Hook callback handler
- `handleCanUseToolTyped(*types.SDKControlPermissionRequest)` - Permission handler
- `handleMCPToolCallTyped(*types.SDKControlMcpToolCallRequest)` - MCP tool call handler

#### Legacy Handlers
Kept existing `map[string]any` handlers for backward compatibility:
- `handleHookCallback(map[string]any)`
- `handleCanUseTool(map[string]any)`
- `handleMCPToolCall(map[string]any)`

#### `handleControlResponse`
- Refactored to parse responses into `types.ControlResponse` struct
- Uses typed `ControlResponseData` for accessing fields
- Eliminates string-based field access with type assertions

### 3. Improved Type Safety

**Before:**
```go
func (q *Query) handleControlRequest(msg map[string]any) {
    request, _ := msg["request"].(map[string]any)
    subtype, _ := request["subtype"].(string)

    switch subtype {
    case "can_use_tool":
        toolName, _ := request["tool_name"].(string)
        input, _ := request["input"].(map[string]any)
        // ...
    }
}
```

**After:**
```go
func (q *Query) handleControlRequest(msg map[string]any) {
    request, _ := msg["request"].(map[string]any)

    typedRequest, parseErr := types.ParseSDKControlRequest(request)
    if parseErr != nil {
        q.sendControlResponse(requestID, nil, fmt.Errorf("failed to parse control request: %w", parseErr))
        return
    }

    switch req := typedRequest.(type) {
    case *types.SDKControlPermissionRequest:
        responseData, err = q.handleCanUseToolTyped(req)
        // Direct access to req.ToolName, req.Input without type assertions
    }
}
```

### 4. Test Coverage

Added comprehensive tests in three files:

#### `sdk/query_control_typed_test.go`
Tests for basic typed struct usage:
- `TestHandleControlResponse_TypedStructs` - Response parsing
- `TestHandleControlResponse_ErrorResponse` - Error response handling
- `TestHandleControlRequest_HookCallback_TypedStruct` - Hook callback with typed request
- `TestHandleControlRequest_CanUseTool_TypedStruct` - Permission request with typed struct
- `TestSendControlResponse_TypedStruct` - Response creation
- `TestSendControlResponse_TypedStruct_Error` - Error response creation

#### `sdk/query_control_refactor_test.go`
Tests for refactored implementation:
- `TestParseControlRequest_UsingTypedStructs` - Verifies typed parsing is used
- `TestHandleControlRequest_ParseError` - Error handling for malformed requests
- `TestHandleCanUseTool_TypedParsing` - Permission callback with correct data extraction
- `TestHandleControlResponse_TypedParsing` - Response parsing preserves types

#### `types/control_test.go`
Added tests for new struct:
- `TestSDKControlMcpToolCallRequest` - MCP tool call request marshaling
- Updated `TestParseSDKControlRequest` with `mcp_tool_call` case

### 5. Helper Files

Created `sdk/test_helpers.go` with shared test utilities:
```go
func stringPtr(s string) *string
```

Fixed `sdk/transport_test.go` missing `time` import.

## Benefits

1. **Type Safety**: Compiler catches field name typos and type mismatches
2. **Better IDE Support**: Auto-completion and go-to-definition for struct fields
3. **Clearer Code**: Intent is explicit through struct types vs. generic maps
4. **Better Error Messages**: Parse errors identify specific issues
5. **Easier Refactoring**: Changes to struct fields are caught at compile time
6. **Documentation**: Struct definitions serve as inline documentation
7. **Backward Compatible**: Legacy handlers preserved for gradual migration

## Backward Compatibility

All changes are backward compatible:
- Existing tests continue to pass
- Legacy `map[string]any` handlers remain available
- JSON serialization/deserialization works identically
- No breaking changes to public API

## Test Results

All tests pass:
```
ok  	github.com/victorarias/claude-agent-sdk-go/internal/mcp	0.164s
ok  	github.com/victorarias/claude-agent-sdk-go/internal/parser	0.300s
ok  	github.com/victorarias/claude-agent-sdk-go/internal/subprocess	3.317s
ok  	github.com/victorarias/claude-agent-sdk-go/sdk	1.840s
ok  	github.com/victorarias/claude-agent-sdk-go/types	0.505s
```

## Files Modified

- `types/messages.go` - Added `SDKControlMcpToolCallRequest`, updated parser
- `sdk/query.go` - Refactored handlers to use typed structs
- `types/control_test.go` - Added tests for new struct
- `sdk/query_control_typed_test.go` - Added typed struct tests
- `sdk/query_control_refactor_test.go` - Added refactoring verification tests
- `sdk/test_helpers.go` - Created shared test utilities
- `sdk/transport_test.go` - Fixed missing import

## Migration Path

For future development:
1. Use typed handlers (`*Typed` functions) for new code
2. Gradually migrate existing code to typed handlers
3. Eventually deprecate and remove legacy `map[string]any` handlers
4. Consider making typed handlers the primary interface
