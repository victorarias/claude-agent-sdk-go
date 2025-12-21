package sdk

import (
	"context"
	"sync"
	"testing"
)

// MockTransport implements Transport for testing.
type MockTransport struct {
	mu           sync.Mutex
	connected    bool
	closed       bool
	written      []string
	inputEnded   bool
	messageChan  chan map[string]any
	connectErr   error
	writeErr     error
	closeErr     error
	endInputErr  error
}

// NewMockTransport creates a new MockTransport.
func NewMockTransport() *MockTransport {
	return &MockTransport{
		messageChan: make(chan map[string]any, 100),
	}
}

func (m *MockTransport) Connect(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.connectErr != nil {
		return m.connectErr
	}
	m.connected = true
	return nil
}

func (m *MockTransport) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closeErr != nil {
		return m.closeErr
	}
	m.closed = true
	m.connected = false
	close(m.messageChan)
	return nil
}

func (m *MockTransport) Write(data string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.writeErr != nil {
		return m.writeErr
	}
	m.written = append(m.written, data)
	return nil
}

func (m *MockTransport) EndInput() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.endInputErr != nil {
		return m.endInputErr
	}
	m.inputEnded = true
	return nil
}

func (m *MockTransport) Messages() <-chan map[string]any {
	return m.messageChan
}

func (m *MockTransport) IsReady() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.connected && !m.closed
}

// SendMessage simulates receiving a message from the CLI.
func (m *MockTransport) SendMessage(msg map[string]any) {
	m.messageChan <- msg
}

// Written returns all data written to the transport.
func (m *MockTransport) Written() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]string, len(m.written))
	copy(result, m.written)
	return result
}

// SetConnectError sets an error to return on Connect.
func (m *MockTransport) SetConnectError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.connectErr = err
}

// SetWriteError sets an error to return on Write.
func (m *MockTransport) SetWriteError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.writeErr = err
}

func TestTransportInterface(t *testing.T) {
	// Verify MockTransport implements Transport
	var _ Transport = (*MockTransport)(nil)
}

func TestMockTransportConnect(t *testing.T) {
	mt := NewMockTransport()

	if mt.IsReady() {
		t.Error("expected transport to not be ready before connect")
	}

	err := mt.Connect(context.Background())
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	if !mt.IsReady() {
		t.Error("expected transport to be ready after connect")
	}
}

func TestMockTransportWrite(t *testing.T) {
	mt := NewMockTransport()
	_ = mt.Connect(context.Background())

	err := mt.Write("test message")
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	written := mt.Written()
	if len(written) != 1 || written[0] != "test message" {
		t.Errorf("expected [\"test message\"], got %v", written)
	}
}

func TestMockTransportMessages(t *testing.T) {
	mt := NewMockTransport()
	_ = mt.Connect(context.Background())

	testMsg := map[string]any{"type": "test", "data": "hello"}
	mt.SendMessage(testMsg)

	select {
	case msg := <-mt.Messages():
		if msg["type"] != "test" || msg["data"] != "hello" {
			t.Errorf("unexpected message: %v", msg)
		}
	default:
		t.Error("expected message on channel")
	}
}

func TestMockTransportClose(t *testing.T) {
	mt := NewMockTransport()
	_ = mt.Connect(context.Background())

	if !mt.IsReady() {
		t.Error("expected transport to be ready")
	}

	err := mt.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	if mt.IsReady() {
		t.Error("expected transport to not be ready after close")
	}
}

func TestMockTransportEndInput(t *testing.T) {
	mt := NewMockTransport()
	_ = mt.Connect(context.Background())

	err := mt.EndInput()
	if err != nil {
		t.Fatalf("EndInput failed: %v", err)
	}

	mt.mu.Lock()
	ended := mt.inputEnded
	mt.mu.Unlock()

	if !ended {
		t.Error("expected inputEnded to be true")
	}
}

func TestMockTransportConnectError(t *testing.T) {
	mt := NewMockTransport()
	mt.SetConnectError(ErrConnection)

	err := mt.Connect(context.Background())
	if err != ErrConnection {
		t.Errorf("expected ErrConnection, got %v", err)
	}
}

func TestMockTransportWriteError(t *testing.T) {
	mt := NewMockTransport()
	_ = mt.Connect(context.Background())
	mt.SetWriteError(ErrClosed)

	err := mt.Write("test")
	if err != ErrClosed {
		t.Errorf("expected ErrClosed, got %v", err)
	}
}
