// Copyright (C) 2025 Claude Agent SDK Go Contributors
// SPDX-License-Identifier: GPL-3.0-only

// Package mcp defines types for the Model Context Protocol (MCP).
//
// This file contains the core type definitions for the MCP protocol, including
// JSON-RPC 2.0 message structures, tool definitions, and protocol constants.
// These types are used by both the handler and transport layers to implement
// the MCP specification.
package mcp

import (
	"encoding/json"

	"github.com/victorarias/claude-agent-sdk-go/types"
)

// MCPProtocolVersion is the MCP protocol version.
const MCPProtocolVersion = "2024-11-05"

// MCPToolDefinition defines a tool exposed via MCP protocol.
type MCPToolDefinition struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	InputSchema map[string]any `json:"inputSchema"`
}

// MCPRequest represents a JSON-RPC 2.0 request.
type MCPRequest struct {
	JSONRPC string         `json:"jsonrpc"`
	ID      any            `json:"id,omitempty"`
	Method  string         `json:"method"`
	Params  map[string]any `json:"params,omitempty"`
}

// MCPResponse represents a JSON-RPC 2.0 response.
type MCPResponse struct {
	JSONRPC string    `json:"jsonrpc"`
	ID      any       `json:"id,omitempty"`
	Result  any       `json:"result,omitempty"`
	Error   *MCPError `json:"error,omitempty"`
}

// MCPError represents a JSON-RPC 2.0 error.
type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// Error implements the error interface.
func (e *MCPError) Error() string {
	return e.Message
}

// Standard MCP error codes
const (
	MCPErrorParseError     = -32700
	MCPErrorInvalidRequest = -32600
	MCPErrorMethodNotFound = -32601
	MCPErrorInvalidParams  = -32602
	MCPErrorInternal       = -32603
)

// MCPNotification represents a JSON-RPC 2.0 notification (no ID).
type MCPNotification struct {
	JSONRPC string         `json:"jsonrpc"`
	Method  string         `json:"method"`
	Params  map[string]any `json:"params,omitempty"`
}

// MCPServerInfo contains server metadata.
type MCPServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// MCPCapabilities defines server capabilities.
type MCPCapabilities struct {
	Tools     *MCPToolsCapability     `json:"tools,omitempty"`
	Resources *MCPResourcesCapability `json:"resources,omitempty"`
	Prompts   *MCPPromptsCapability   `json:"prompts,omitempty"`
}

// MCPToolsCapability defines tools capability.
type MCPToolsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// MCPResourcesCapability defines resources capability.
type MCPResourcesCapability struct {
	Subscribe   bool `json:"subscribe,omitempty"`
	ListChanged bool `json:"listChanged,omitempty"`
}

// MCPPromptsCapability defines prompts capability.
type MCPPromptsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// MCPInitializeResult is the result of initialize request.
type MCPInitializeResult struct {
	ProtocolVersion string           `json:"protocolVersion"`
	Capabilities    *MCPCapabilities `json:"capabilities"`
	ServerInfo      MCPServerInfo    `json:"serverInfo"`
}

// MCPToolsListResult is the result of tools/list request.
type MCPToolsListResult struct {
	Tools []MCPToolDefinition `json:"tools"`
}

// MCPToolCallResult is the result of tools/call request.
type MCPToolCallResult struct {
	Content []types.MCPContent `json:"content"`
	IsError bool               `json:"isError,omitempty"`
}

// NewMCPError creates a new MCP error.
func NewMCPError(code int, message string, data any) *MCPError {
	return &MCPError{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

// NewMCPResponse creates a successful response.
func NewMCPResponse(id any, result any) *MCPResponse {
	return &MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
}

// NewMCPErrorResponse creates an error response.
func NewMCPErrorResponse(id any, err *MCPError) *MCPResponse {
	return &MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error:   err,
	}
}

// ParseMCPRequest parses a JSON-RPC request from bytes.
func ParseMCPRequest(data []byte) (*MCPRequest, error) {
	var req MCPRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, err
	}
	return &req, nil
}

// MarshalMCPResponse serializes a response to JSON.
func MarshalMCPResponse(resp *MCPResponse) ([]byte, error) {
	return json.Marshal(resp)
}

// NewMCPTextContent creates a text content item.
func NewMCPTextContent(text string) types.MCPContent {
	return types.MCPContent{
		Type: "text",
		Text: text,
	}
}
