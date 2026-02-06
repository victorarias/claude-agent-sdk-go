// Copyright (C) 2025 Claude Agent SDK Go Contributors
// SPDX-License-Identifier: GPL-3.0-only

package types

import (
	"context"
)

// Transport defines the interface for communicating with Claude.
type Transport interface {
	// Connect establishes the connection.
	Connect(ctx context.Context) error

	// Close terminates the connection and cleans up resources.
	Close() error

	// Write sends data to the CLI.
	Write(data string) error

	// EndInput signals that no more input will be sent.
	EndInput() error

	// Messages returns a channel of parsed JSON messages from the CLI.
	Messages() <-chan map[string]any

	// IsReady returns true if the transport is connected and ready.
	IsReady() bool
}

// ErrorTransport is an optional transport extension for asynchronous transport errors.
//
// Implement this on custom transports to propagate transport-level errors (e.g. parser
// or subprocess failures) to Query and Client error streams.
type ErrorTransport interface {
	Errors() <-chan error
}

// LegacyTransport represents the transitional interface shape where Transport also
// included Errors(). It is kept as a compatibility shim.
//
// Deprecated: implement Transport and optionally ErrorTransport instead.
type LegacyTransport interface {
	Transport
	ErrorTransport
}
