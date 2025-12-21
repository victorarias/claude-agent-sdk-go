package sdk

import (
	"sync"
)

// MockTransport is a test double for Transport.
type MockTransport struct {
	messages chan map[string]any
	errors   chan error
	written  []string
	writeMu  sync.Mutex
	closed   bool
}

// NewMockTransport creates a new MockTransport.
func NewMockTransport() *MockTransport {
	return &MockTransport{
		messages: make(chan map[string]any, 100),
		errors:   make(chan error, 1),
		written:  make([]string, 0),
	}
}

// Messages returns the channel for receiving messages.
func (m *MockTransport) Messages() <-chan map[string]any {
	return m.messages
}

// Errors returns the channel for receiving errors.
func (m *MockTransport) Errors() <-chan error {
	return m.errors
}

// Write records data written to the transport.
func (m *MockTransport) Write(data string) error {
	m.writeMu.Lock()
	defer m.writeMu.Unlock()
	m.written = append(m.written, data)
	return nil
}

// WriteJSON writes a JSON object (simplified for tests).
func (m *MockTransport) WriteJSON(obj any) error {
	return nil // Simplified for tests
}

// Close closes the mock transport.
func (m *MockTransport) Close() error {
	if !m.closed {
		m.closed = true
		close(m.messages)
	}
	return nil
}

// IsReady returns true if the transport is ready.
func (m *MockTransport) IsReady() bool {
	return !m.closed
}

// Test helpers

// Written returns a copy of all data written to the transport.
func (m *MockTransport) Written() []string {
	m.writeMu.Lock()
	defer m.writeMu.Unlock()
	return append([]string{}, m.written...)
}

// SendMessage simulates receiving a message from the CLI.
func (m *MockTransport) SendMessage(msg map[string]any) {
	m.messages <- msg
}

// SendError simulates receiving an error from the CLI.
func (m *MockTransport) SendError(err error) {
	m.errors <- err
}

// ClearWritten clears the written data.
func (m *MockTransport) ClearWritten() {
	m.writeMu.Lock()
	defer m.writeMu.Unlock()
	m.written = make([]string, 0)
}
