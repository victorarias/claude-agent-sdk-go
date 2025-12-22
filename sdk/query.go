// Copyright (C) 2025 Claude Agent SDK Go Contributors
// SPDX-License-Identifier: GPL-3.0-only

package sdk

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/victorarias/claude-agent-sdk-go/internal/parser"
	"github.com/victorarias/claude-agent-sdk-go/types"
)

// DefaultStreamCloseTimeout is the default timeout for waiting for first result
// before closing stdin when hooks or MCP servers are active.
const DefaultStreamCloseTimeout = 60 * time.Second

// MessageChannelBuffer is the buffer size for the parsed messages channel.
// A buffer of 100 provides sufficient capacity to handle message bursts without
// blocking the message router, while preventing unbounded memory growth.
const MessageChannelBuffer = 100

// RawMessageChannelBuffer is the buffer size for the raw messages channel.
// A buffer of 100 matches the parsed messages channel capacity, ensuring
// consistent buffering behavior for both message processing paths.
const RawMessageChannelBuffer = 100

// Query handles the bidirectional control protocol.
type Query struct {
	transport types.Transport
	streaming bool

	// Control protocol state
	pendingRequests map[string]chan map[string]any
	pendingMu       sync.Mutex
	requestCounter  atomic.Uint64

	// Hook callbacks
	hookCallbacks  map[string]types.HookCallback
	nextCallbackID atomic.Uint64
	hookMu         sync.RWMutex

	// Permission callback
	canUseTool types.CanUseToolCallback

	// MCP server registry
	mcpServers   map[string]*types.MCPServer
	mcpServersMu sync.RWMutex

	// Message channels
	messages    chan types.Message  // Parsed messages
	rawMessages chan map[string]any // Raw messages for custom handling
	errors      chan error

	// Result tracking
	resultReceived  atomic.Bool
	lastResult      *types.ResultMessage
	resultMu        sync.RWMutex
	firstResultChan chan struct{} // Closed when first result is received
	firstResultOnce sync.Once     // Ensures channel is closed only once

	// Stream close timeout for waiting for first result
	streamCloseTimeout time.Duration

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	closed atomic.Bool

	// Initialization result
	initResult map[string]any
	initMu     sync.RWMutex
}

// NewQuery creates a new Query.
func NewQuery(transport types.Transport, streaming bool) *Query {
	return &Query{
		transport:          transport,
		streaming:          streaming,
		pendingRequests:    make(map[string]chan map[string]any),
		hookCallbacks:      make(map[string]types.HookCallback),
		mcpServers:         make(map[string]*types.MCPServer),
		messages:           make(chan types.Message, MessageChannelBuffer),
		rawMessages:        make(chan map[string]any, RawMessageChannelBuffer),
		errors:             make(chan error, 1),
		firstResultChan:    make(chan struct{}),
		streamCloseTimeout: DefaultStreamCloseTimeout,
	}
}

// Start begins processing messages from the transport.
func (q *Query) Start(ctx context.Context) error {
	q.ctx, q.cancel = context.WithCancel(ctx)

	// Start message router
	q.wg.Add(1)
	go q.routeMessages()

	return nil
}

// Messages returns the channel of parsed SDK messages.
func (q *Query) Messages() <-chan types.Message {
	return q.messages
}

// RawMessages returns the channel of raw messages for custom handling.
func (q *Query) RawMessages() <-chan map[string]any {
	return q.rawMessages
}

// Errors returns the channel of errors.
func (q *Query) Errors() <-chan error {
	return q.errors
}

// ResultReceived returns true if a result message has been received.
func (q *Query) ResultReceived() bool {
	return q.resultReceived.Load()
}

// LastResult returns the last result message received.
func (q *Query) LastResult() *types.ResultMessage {
	q.resultMu.RLock()
	defer q.resultMu.RUnlock()
	return q.lastResult
}

// Close stops the query.
func (q *Query) Close() error {
	if q.closed.Swap(true) {
		return nil // Already closed
	}

	if q.cancel != nil {
		q.cancel()
	}

	q.wg.Wait()

	// Close channels after goroutines finish
	close(q.messages)
	close(q.rawMessages)

	return nil
}

// routeMessages reads from transport and routes control vs SDK messages.
func (q *Query) routeMessages() {
	defer q.wg.Done()

	for {
		select {
		case <-q.ctx.Done():
			return
		case raw, ok := <-q.transport.Messages():
			if !ok {
				return
			}

			msgType, _ := raw["type"].(string)

			switch msgType {
			case "control_response":
				q.handleControlResponse(raw)
			case "control_request":
				go q.handleControlRequest(raw)
			case "control_cancel_request":
				q.handleCancelRequest(raw)
			default:
				// Parse and route regular messages
				q.handleSDKMessage(raw)
			}
		}
	}
}

// handleSDKMessage parses and routes an SDK message.
func (q *Query) handleSDKMessage(raw map[string]any) {
	// Send to raw channel for custom handling
	select {
	case q.rawMessages <- raw:
	default:
		// Drop if full
	}

	// Parse message
	msg, err := parser.ParseMessage(raw)
	if err != nil {
		// Still send to channel for unknown message types
		select {
		case q.errors <- err:
		default:
		}
		return
	}

	// Track result messages
	if result, ok := msg.(*types.ResultMessage); ok {
		q.resultMu.Lock()
		q.lastResult = result
		q.resultMu.Unlock()
		q.resultReceived.Store(true)
		// Signal first result received (only closes once)
		q.firstResultOnce.Do(func() {
			close(q.firstResultChan)
		})
	}

	// Send parsed message
	select {
	case q.messages <- msg:
	case <-q.ctx.Done():
	}
}

// handleControlResponse routes a control response to the waiting request.
func (q *Query) handleControlResponse(msg map[string]any) {
	// Parse the response into a typed struct
	responseBytes, err := json.Marshal(msg)
	if err != nil {
		return // Cannot parse, silently drop
	}

	var typedResponse types.ControlResponse
	if err := json.Unmarshal(responseBytes, &typedResponse); err != nil {
		return // Cannot parse, silently drop
	}

	requestID := typedResponse.Response.RequestID
	if requestID == "" {
		return
	}

	q.pendingMu.Lock()
	respChan, exists := q.pendingRequests[requestID]
	q.pendingMu.Unlock()

	if !exists {
		return
	}

	// Check for error
	if typedResponse.Response.Subtype == "error" {
		select {
		case respChan <- map[string]any{"error": typedResponse.Response.Error}:
		default:
		}
		return
	}

	// Send success response
	respData := typedResponse.Response.Response
	if respData == nil {
		respData = make(map[string]any)
	}

	select {
	case respChan <- respData:
	default:
	}
}

// handleControlRequest handles incoming control requests from CLI.
func (q *Query) handleControlRequest(msg map[string]any) {
	requestID, _ := msg["request_id"].(string)
	request, _ := msg["request"].(map[string]any)
	if request == nil {
		return
	}

	// Parse the request into a typed struct
	typedRequest, parseErr := types.ParseSDKControlRequest(request)
	if parseErr != nil {
		q.sendControlResponse(requestID, nil, fmt.Errorf("failed to parse control request: %w", parseErr))
		return
	}

	var responseData map[string]any
	var err error

	// Handle based on typed request
	switch req := typedRequest.(type) {
	case *types.SDKControlPermissionRequest:
		responseData, err = q.handleCanUseToolTyped(req)
	case *types.SDKHookCallbackRequest:
		responseData, err = q.handleHookCallbackTyped(req)
	case *types.SDKControlMcpToolCallRequest:
		responseData, err = q.handleMCPToolCallTyped(req)
	case *types.SDKControlMcpMessageRequest:
		var mcpResponse any
		mcpResponse, err = q.handleMCPMessage(req.ServerName, req.Message.(map[string]any))
		if err == nil {
			// Wrap the MCP response as expected by the control protocol
			responseData = map[string]any{"mcp_response": mcpResponse}
		}
	case *types.SDKControlInterruptRequest:
		// Interrupt is handled via sendControlRequest, not incoming
		err = fmt.Errorf("interrupt request not expected as incoming request")
	case *types.SDKControlInitializeRequest:
		// Initialize is handled via sendControlRequest, not incoming
		err = fmt.Errorf("initialize request not expected as incoming request")
	case *types.SDKControlSetPermissionModeRequest:
		// Set permission mode is handled via sendControlRequest, not incoming
		err = fmt.Errorf("set permission mode request not expected as incoming request")
	case *types.SDKControlRewindFilesRequest:
		// Rewind files is handled via sendControlRequest, not incoming
		err = fmt.Errorf("rewind files request not expected as incoming request")
	default:
		err = fmt.Errorf("unsupported control request type: %T", typedRequest)
	}

	// Send response
	q.sendControlResponse(requestID, responseData, err)
}

// handleHookCallbackTyped invokes a registered hook callback using typed request.
func (q *Query) handleHookCallbackTyped(req *types.SDKHookCallbackRequest) (map[string]any, error) {
	q.hookMu.RLock()
	callback, exists := q.hookCallbacks[req.CallbackID]
	q.hookMu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("hook callback not found: %s", req.CallbackID)
	}

	output, err := callback(req.Input, req.ToolUseID, &types.HookContext{})
	if err != nil {
		return nil, err
	}

	// Convert HookOutput to response
	return q.hookOutputToResponse(output), nil
}

// handleHookCallback invokes a registered hook callback (legacy, kept for backward compatibility).
func (q *Query) handleHookCallback(request map[string]any) (map[string]any, error) {
	callbackID, _ := request["callback_id"].(string)
	input := request["input"]

	var toolUseID *string
	if id, ok := request["tool_use_id"].(string); ok {
		toolUseID = &id
	}

	q.hookMu.RLock()
	callback, exists := q.hookCallbacks[callbackID]
	q.hookMu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("hook callback not found: %s", callbackID)
	}

	output, err := callback(input, toolUseID, &types.HookContext{})
	if err != nil {
		return nil, err
	}

	// Convert HookOutput to response
	return q.hookOutputToResponse(output), nil
}

// hookOutputToResponse converts a HookOutput to a response map.
func (q *Query) hookOutputToResponse(output *types.HookOutput) map[string]any {
	if output == nil {
		return make(map[string]any)
	}

	result := make(map[string]any)

	if output.Continue != nil && *output.Continue {
		result["continue"] = true
	}
	if output.SuppressOutput {
		result["suppressOutput"] = true
	}
	if output.StopReason != "" {
		result["stopReason"] = output.StopReason
	}
	if output.Decision != "" {
		result["decision"] = output.Decision
	}
	if output.SystemMessage != "" {
		result["systemMessage"] = output.SystemMessage
	}
	if output.Reason != "" {
		result["reason"] = output.Reason
	}
	if output.HookSpecific != nil {
		result["hookSpecificOutput"] = output.HookSpecific
	}

	return result
}

// handleCanUseToolTyped handles tool permission requests using typed request.
func (q *Query) handleCanUseToolTyped(req *types.SDKControlPermissionRequest) (map[string]any, error) {
	if q.canUseTool == nil {
		// Default: allow all
		return map[string]any{"behavior": "allow"}, nil
	}

	ctx := &types.ToolPermissionContext{
		Suggestions: req.PermissionSuggestions,
		BlockedPath: req.BlockedPath,
	}

	result, err := q.canUseTool(req.ToolName, req.Input, ctx)
	if err != nil {
		return nil, err
	}

	return q.permissionResultToResponse(result)
}

// handleCanUseTool handles tool permission requests (legacy, kept for backward compatibility).
func (q *Query) handleCanUseTool(request map[string]any) (map[string]any, error) {
	if q.canUseTool == nil {
		// Default: allow all
		return map[string]any{"behavior": "allow"}, nil
	}

	toolName, _ := request["tool_name"].(string)
	input, _ := request["input"].(map[string]any)
	suggestions, _ := request["permission_suggestions"].([]any)

	ctx := &types.ToolPermissionContext{
		Suggestions: q.parsePermissionSuggestions(suggestions),
	}

	result, err := q.canUseTool(toolName, input, ctx)
	if err != nil {
		return nil, err
	}

	return q.permissionResultToResponse(result)
}

// parsePermissionSuggestions converts raw suggestions to PermissionUpdate slice.
func (q *Query) parsePermissionSuggestions(raw []any) []types.PermissionUpdate {
	if raw == nil {
		return nil
	}

	var updates []types.PermissionUpdate
	for _, item := range raw {
		if m, ok := item.(map[string]any); ok {
			update := types.PermissionUpdate{
				Type: types.PermissionUpdateType(getString(m, "type")),
			}
			updates = append(updates, update)
		}
	}
	return updates
}

// permissionResultToResponse converts a permission result to a response map.
func (q *Query) permissionResultToResponse(result types.PermissionResult) (map[string]any, error) {
	switch r := result.(type) {
	case *types.PermissionResultAllow:
		resp := map[string]any{
			"behavior": "allow",
		}
		if r.UpdatedInput != nil {
			resp["updatedInput"] = r.UpdatedInput
		}
		if len(r.UpdatedPermissions) > 0 {
			resp["permissionUpdates"] = r.UpdatedPermissions
		}
		return resp, nil

	case *types.PermissionResultDeny:
		resp := map[string]any{
			"behavior": "deny",
			"message":  r.Message,
		}
		if r.Interrupt {
			resp["interrupt"] = true
		}
		return resp, nil

	default:
		return nil, fmt.Errorf("invalid permission result type: %T", result)
	}
}

// sendControlResponse sends a control response.
func (q *Query) sendControlResponse(requestID string, data map[string]any, err error) {
	response := map[string]any{
		"type": "control_response",
	}

	if err != nil {
		response["response"] = map[string]any{
			"subtype":    "error",
			"request_id": requestID,
			"error":      err.Error(),
		}
	} else {
		response["response"] = map[string]any{
			"subtype":    "success",
			"request_id": requestID,
			"response":   data,
		}
	}

	responseData, marshalErr := json.Marshal(response)
	if marshalErr != nil {
		// Cannot marshal response - silently return to avoid writing invalid data
		return
	}
	q.transport.Write(string(responseData))
}

// SetCanUseTool sets the tool permission callback.
func (q *Query) SetCanUseTool(callback types.CanUseToolCallback) {
	q.canUseTool = callback
}

// SendMessage sends a single message to the CLI.
func (q *Query) SendMessage(msg map[string]any) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	return q.transport.Write(string(data))
}

// SendUserMessage sends a user message.
func (q *Query) SendUserMessage(content string, sessionID string) error {
	msg := map[string]any{
		"type": "user",
		"message": map[string]any{
			"role":    "user",
			"content": content,
		},
		"parent_tool_use_id": nil,
		"session_id":         sessionID,
	}
	return q.SendMessage(msg)
}

// StreamInput streams messages from a channel to the CLI.
func (q *Query) StreamInput(input <-chan map[string]any) error {
	for {
		select {
		case <-q.ctx.Done():
			return q.ctx.Err()
		case msg, ok := <-input:
			if !ok {
				return nil // Channel closed
			}
			if err := q.SendMessage(msg); err != nil {
				return err
			}
		}
	}
}

// StreamInputWithWait streams messages and waits for first result if hooks/MCP are active.
// This implements the stream closure coordination from the Python SDK - when hooks or MCP
// servers are active, we must wait for the first result before allowing stdin to be closed,
// otherwise bidirectional control protocol communication may be interrupted.
func (q *Query) StreamInputWithWait(input <-chan map[string]any) error {
	// Stream all messages from input
	if err := q.StreamInput(input); err != nil {
		return err
	}

	// If hooks or MCP servers are active, wait for first result before returning
	// This allows the control protocol to continue functioning
	if q.HasActiveHooksOrMCP() {
		select {
		case <-q.firstResultChan:
			// First result received, safe to proceed
		case <-time.After(q.streamCloseTimeout):
			// Timeout - proceed anyway to avoid hanging forever
		case <-q.ctx.Done():
			return q.ctx.Err()
		}
	}

	return nil
}

// SetStreamCloseTimeout sets the timeout for waiting for first result
// before closing stdin when hooks or MCP servers are active.
func (q *Query) SetStreamCloseTimeout(timeout time.Duration) {
	q.streamCloseTimeout = timeout
}

// WaitForFirstResult returns a channel that is closed when the first result is received.
func (q *Query) WaitForFirstResult() <-chan struct{} {
	return q.firstResultChan
}

// HasActiveHooksOrMCP returns true if there are any registered hook callbacks or MCP servers.
func (q *Query) HasActiveHooksOrMCP() bool {
	q.hookMu.RLock()
	hasHooks := len(q.hookCallbacks) > 0
	q.hookMu.RUnlock()

	q.mcpServersMu.RLock()
	hasMCP := len(q.mcpServers) > 0
	q.mcpServersMu.RUnlock()

	return hasHooks || hasMCP
}

// RegisterMCPServer registers an MCP server with the query.
func (q *Query) RegisterMCPServer(server *types.MCPServer) {
	q.mcpServersMu.Lock()
	defer q.mcpServersMu.Unlock()
	q.mcpServers[server.Name] = server
}

// UnregisterMCPServer removes an MCP server.
func (q *Query) UnregisterMCPServer(name string) {
	q.mcpServersMu.Lock()
	defer q.mcpServersMu.Unlock()
	delete(q.mcpServers, name)
}

// handleMCPToolCallTyped handles MCP tool call requests using typed request.
func (q *Query) handleMCPToolCallTyped(req *types.SDKControlMcpToolCallRequest) (map[string]any, error) {
	q.mcpServersMu.RLock()
	server, exists := q.mcpServers[req.ServerName]
	q.mcpServersMu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("MCP server not found: %s", req.ServerName)
	}

	result, err := server.CallTool(req.ToolName, req.Input)
	if err != nil {
		return nil, err
	}

	return map[string]any{"result": result}, nil
}

// handleMCPToolCall handles MCP tool call requests (legacy, kept for backward compatibility).
func (q *Query) handleMCPToolCall(request map[string]any) (map[string]any, error) {
	serverName, _ := request["server_name"].(string)
	toolName, _ := request["tool_name"].(string)
	input, _ := request["input"].(map[string]any)

	q.mcpServersMu.RLock()
	server, exists := q.mcpServers[serverName]
	q.mcpServersMu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("MCP server not found: %s", serverName)
	}

	result, err := server.CallTool(toolName, input)
	if err != nil {
		return nil, err
	}

	return map[string]any{"result": result}, nil
}

// handleMCPMessage handles MCP JSONRPC messages for SDK-hosted MCP servers.
// This bridges JSONRPC messages from the CLI to the in-process MCP server.
func (q *Query) handleMCPMessage(serverName string, message map[string]any) (any, error) {
	q.mcpServersMu.RLock()
	server, exists := q.mcpServers[serverName]
	q.mcpServersMu.RUnlock()

	if !exists {
		// Return JSONRPC error response for server not found
		return map[string]any{
			"jsonrpc": "2.0",
			"id":      message["id"],
			"error": map[string]any{
				"code":    -32601,
				"message": fmt.Sprintf("Server '%s' not found", serverName),
			},
		}, nil
	}

	method, _ := message["method"].(string)
	params, _ := message["params"].(map[string]any)
	if params == nil {
		params = make(map[string]any)
	}
	msgID := message["id"]

	// Handle notifications (no ID)
	if msgID == nil {
		switch method {
		case "notifications/initialized":
			// Acknowledged, no response needed
			return nil, nil
		case "notifications/cancelled":
			// TODO: Implement cancellation
			return nil, nil
		}
		return nil, nil
	}

	// Handle requests
	switch method {
	case "initialize":
		return map[string]any{
			"jsonrpc": "2.0",
			"id":      msgID,
			"result": map[string]any{
				"protocolVersion": "2024-11-05",
				"capabilities": map[string]any{
					"tools": map[string]any{},
				},
				"serverInfo": map[string]any{
					"name":    server.Name,
					"version": server.Version,
				},
			},
		}, nil

	case "tools/list":
		tools := make([]any, 0, len(server.Tools))
		for _, tool := range server.Tools {
			schema := tool.Schema
			if schema == nil {
				schema = map[string]any{"type": "object"}
			}
			tools = append(tools, map[string]any{
				"name":        tool.Name,
				"description": tool.Description,
				"inputSchema": schema,
			})
		}
		return map[string]any{
			"jsonrpc": "2.0",
			"id":      msgID,
			"result": map[string]any{
				"tools": tools,
			},
		}, nil

	case "tools/call":
		toolName, _ := params["name"].(string)
		args, _ := params["arguments"].(map[string]any)
		if args == nil {
			args = make(map[string]any)
		}

		result, err := server.CallTool(toolName, args)
		if err != nil {
			return map[string]any{
				"jsonrpc": "2.0",
				"id":      msgID,
				"error": map[string]any{
					"code":    -32603,
					"message": err.Error(),
				},
			}, nil
		}

		// Convert result to response format
		content := make([]any, len(result.Content))
		for i, item := range result.Content {
			contentItem := map[string]any{
				"type": item.Type,
			}
			if item.Text != "" {
				contentItem["text"] = item.Text
			}
			if item.Data != "" {
				contentItem["data"] = item.Data
			}
			if item.MimeType != "" {
				contentItem["mimeType"] = item.MimeType
			}
			content[i] = contentItem
		}

		responseData := map[string]any{
			"content": content,
		}
		if result.IsError {
			responseData["isError"] = true
		}

		return map[string]any{
			"jsonrpc": "2.0",
			"id":      msgID,
			"result":  responseData,
		}, nil

	case "ping":
		return map[string]any{
			"jsonrpc": "2.0",
			"id":      msgID,
			"result":  map[string]any{},
		}, nil

	default:
		return map[string]any{
			"jsonrpc": "2.0",
			"id":      msgID,
			"error": map[string]any{
				"code":    -32601,
				"message": fmt.Sprintf("Method '%s' not found", method),
			},
		}, nil
	}
}

// handleCancelRequest handles a request cancellation.
func (q *Query) handleCancelRequest(raw map[string]any) {
	requestID, _ := raw["request_id"].(string)
	if requestID == "" {
		return
	}

	q.pendingMu.Lock()
	if respChan, exists := q.pendingRequests[requestID]; exists {
		// Send cancellation signal
		select {
		case respChan <- map[string]any{"cancelled": true}:
		default:
		}
	}
	q.pendingMu.Unlock()
}

// sendControlRequest sends a control request and waits for response.
func (q *Query) sendControlRequest(request map[string]any, timeout time.Duration) (map[string]any, error) {
	if !q.streaming {
		return nil, fmt.Errorf("control requests require streaming mode")
	}

	// Generate request ID with random hex to prevent collisions
	// Matches Python SDK format: req_{counter}_{random_hex}
	id := q.requestCounter.Add(1)
	randomBytes := make([]byte, 4)
	rand.Read(randomBytes)
	requestID := fmt.Sprintf("req_%d_%s", id, hex.EncodeToString(randomBytes))

	// Create response channel
	respChan := make(chan map[string]any, 1)
	q.pendingMu.Lock()
	q.pendingRequests[requestID] = respChan
	q.pendingMu.Unlock()

	defer func() {
		q.pendingMu.Lock()
		delete(q.pendingRequests, requestID)
		q.pendingMu.Unlock()
	}()

	// Build and send request
	controlReq := map[string]any{
		"type":       "control_request",
		"request_id": requestID,
		"request":    request,
	}

	data, err := json.Marshal(controlReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	if err := q.transport.Write(string(data)); err != nil {
		return nil, fmt.Errorf("failed to write request: %w", err)
	}

	// Wait for response
	select {
	case resp := <-respChan:
		if errMsg, ok := resp["error"].(string); ok {
			return nil, fmt.Errorf("control request error: %s", errMsg)
		}
		if resp["cancelled"] == true {
			return nil, fmt.Errorf("control request cancelled")
		}
		return resp, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("control request timeout: %v", request["subtype"])
	case <-q.ctx.Done():
		return nil, q.ctx.Err()
	}
}

// Initialize sends the initialization request to the CLI.
func (q *Query) Initialize(hooks map[types.HookEvent][]types.HookMatcher) (map[string]any, error) {
	if !q.streaming {
		return nil, nil
	}

	// Build hooks configuration
	hooksConfig := q.buildHooksConfig(hooks)

	request := map[string]any{
		"subtype": "initialize",
	}
	if len(hooksConfig) > 0 {
		request["hooks"] = hooksConfig
	}

	result, err := q.sendControlRequest(request, 60*time.Second)
	if err != nil {
		return nil, err
	}

	q.initMu.Lock()
	q.initResult = result
	q.initMu.Unlock()

	return result, nil
}

// buildHooksConfig builds the hooks configuration for initialization.
func (q *Query) buildHooksConfig(hooks map[types.HookEvent][]types.HookMatcher) map[string]any {
	if hooks == nil {
		return nil
	}

	config := make(map[string]any)

	for event, matchers := range hooks {
		if len(matchers) == 0 {
			continue
		}

		var matcherConfigs []map[string]any
		for _, matcher := range matchers {
			callbackIDs := make([]string, len(matcher.Hooks))
			for i, callback := range matcher.Hooks {
				callbackID := fmt.Sprintf("hook_%d", q.nextCallbackID.Add(1))
				q.hookMu.Lock()
				q.hookCallbacks[callbackID] = callback
				q.hookMu.Unlock()
				callbackIDs[i] = callbackID
			}

			matcherConfig := map[string]any{
				"matcher":         matcher.Matcher,
				"hookCallbackIds": callbackIDs,
			}
			if matcher.Timeout != nil && *matcher.Timeout > 0 {
				matcherConfig["timeout"] = *matcher.Timeout
			}
			matcherConfigs = append(matcherConfigs, matcherConfig)
		}
		config[string(event)] = matcherConfigs
	}

	return config
}

// InitResult returns the initialization result.
func (q *Query) InitResult() map[string]any {
	q.initMu.RLock()
	defer q.initMu.RUnlock()
	return q.initResult
}

// Interrupt sends an interrupt signal to the CLI.
func (q *Query) Interrupt() error {
	_, err := q.sendControlRequest(map[string]any{
		"subtype": "interrupt",
	}, 30*time.Second)
	return err
}

// SetPermissionMode changes the permission mode.
func (q *Query) SetPermissionMode(mode types.PermissionMode) error {
	_, err := q.sendControlRequest(map[string]any{
		"subtype": "set_permission_mode",
		"mode":    string(mode),
	}, 30*time.Second)
	return err
}

// SetModel changes the AI model.
func (q *Query) SetModel(model string) error {
	_, err := q.sendControlRequest(map[string]any{
		"subtype": "set_model",
		"model":   model,
	}, 30*time.Second)
	return err
}

// RewindFiles rewinds tracked files to a specific user message.
func (q *Query) RewindFiles(userMessageID string) error {
	_, err := q.sendControlRequest(map[string]any{
		"subtype":         "rewind_files",
		"user_message_id": userMessageID,
	}, 30*time.Second)
	return err
}

// getString safely extracts a string from a map.
func getString(m map[string]any, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}
