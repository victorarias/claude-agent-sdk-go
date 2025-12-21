package sdk

import (
	"context"
	"errors"
	"testing"

	"github.com/victorarias/claude-agent-sdk-go/types"
)

// MockTransport is a minimal mock for testing WithClient
type mockTransportWithClient struct {
	connected      bool
	connectErr     error
	closeCallCount int
	closeShouldErr bool
}

func (m *mockTransportWithClient) Connect(ctx context.Context) error {
	if m.connectErr != nil {
		return m.connectErr
	}
	m.connected = true
	return nil
}

func (m *mockTransportWithClient) Close() error {
	m.closeCallCount++
	m.connected = false
	if m.closeShouldErr {
		return errors.New("close error")
	}
	return nil
}

func (m *mockTransportWithClient) Messages() <-chan map[string]any {
	ch := make(chan map[string]any)
	close(ch)
	return ch
}

func (m *mockTransportWithClient) Send(msg map[string]any) error {
	return nil
}

// TestWithClient_Success tests that WithClient creates client, runs function, and disconnects
func TestWithClient_Success(t *testing.T) {
	mock := &mockTransportWithClient{}
	opts := []types.Option{
		types.WithCustomTransport(mock),
	}

	functionCalled := false
	err := WithClient(context.Background(), opts, func(c *Client) error {
		functionCalled = true
		if !c.IsConnected() {
			t.Error("expected client to be connected")
		}
		return nil
	})

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if !functionCalled {
		t.Error("expected function to be called")
	}
	if mock.closeCallCount != 1 {
		t.Errorf("expected Close to be called once, got %d calls", mock.closeCallCount)
	}
	if mock.connected {
		t.Error("expected transport to be disconnected after WithClient returns")
	}
}

// TestWithClient_FunctionError tests that disconnect is called even if function returns error
func TestWithClient_FunctionError(t *testing.T) {
	mock := &mockTransportWithClient{}
	opts := []types.Option{
		types.WithCustomTransport(mock),
	}

	expectedErr := errors.New("function error")
	err := WithClient(context.Background(), opts, func(c *Client) error {
		return expectedErr
	})

	if err != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
	if mock.closeCallCount != 1 {
		t.Errorf("expected Close to be called once even on error, got %d calls", mock.closeCallCount)
	}
	if mock.connected {
		t.Error("expected transport to be disconnected even after function error")
	}
}

// TestWithClient_FunctionPanic tests that disconnect is called even if function panics
func TestWithClient_FunctionPanic(t *testing.T) {
	mock := &mockTransportWithClient{}
	opts := []types.Option{
		types.WithCustomTransport(mock),
	}

	expectedPanic := "intentional panic"

	// We need to recover from the panic that WithClient should re-raise
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic to be re-raised, but no panic occurred")
		}
		if r != expectedPanic {
			t.Errorf("expected panic value %v, got %v", expectedPanic, r)
		}

		// Verify cleanup happened before panic was re-raised
		if mock.closeCallCount != 1 {
			t.Errorf("expected Close to be called once even on panic, got %d calls", mock.closeCallCount)
		}
		if mock.connected {
			t.Error("expected transport to be disconnected even after panic")
		}
	}()

	WithClient(context.Background(), opts, func(c *Client) error {
		panic(expectedPanic)
	})

	// Should not reach here
	t.Fatal("expected function to panic, but it didn't")
}

// TestWithClient_PanicReraise tests that original panic is re-raised after cleanup
func TestWithClient_PanicReraise(t *testing.T) {
	mock := &mockTransportWithClient{}
	opts := []types.Option{
		types.WithCustomTransport(mock),
	}

	type customPanic struct {
		message string
		code    int
	}

	expectedPanic := customPanic{message: "custom panic", code: 42}

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic to be re-raised")
		}

		cp, ok := r.(customPanic)
		if !ok {
			t.Errorf("expected customPanic type, got %T", r)
		}
		if cp != expectedPanic {
			t.Errorf("expected panic %+v, got %+v", expectedPanic, cp)
		}
	}()

	WithClient(context.Background(), opts, func(c *Client) error {
		panic(expectedPanic)
	})
}

// TestWithClient_ConnectError tests that function is not called if connect fails
func TestWithClient_ConnectError(t *testing.T) {
	connectErr := errors.New("connect failed")
	mock := &mockTransportWithClient{
		connectErr: connectErr,
	}
	opts := []types.Option{
		types.WithCustomTransport(mock),
	}

	functionCalled := false
	err := WithClient(context.Background(), opts, func(c *Client) error {
		functionCalled = true
		return nil
	})

	if err != connectErr {
		t.Errorf("expected connect error %v, got %v", connectErr, err)
	}
	if functionCalled {
		t.Error("expected function not to be called when connect fails")
	}
	// Close should not be called if connect fails
	if mock.closeCallCount != 0 {
		t.Errorf("expected Close not to be called when connect fails, got %d calls", mock.closeCallCount)
	}
}

// TestWithClient_DisconnectError tests error aggregation when both function and disconnect fail
func TestWithClient_DisconnectError(t *testing.T) {
	mock := &mockTransportWithClient{
		closeShouldErr: true,
	}
	opts := []types.Option{
		types.WithCustomTransport(mock),
	}

	functionErr := errors.New("function error")
	err := WithClient(context.Background(), opts, func(c *Client) error {
		return functionErr
	})

	// The function error should be returned (primary error)
	// Note: The current implementation in client.go doesn't aggregate errors,
	// it just returns the function error. This test documents that behavior.
	if err != functionErr {
		t.Errorf("expected function error %v, got %v", functionErr, err)
	}

	// Close should still be called
	if mock.closeCallCount != 1 {
		t.Errorf("expected Close to be called, got %d calls", mock.closeCallCount)
	}
}

// TestWithClient_DisconnectErrorOnly tests when only disconnect fails
func TestWithClient_DisconnectErrorOnly(t *testing.T) {
	mock := &mockTransportWithClient{
		closeShouldErr: true,
	}
	opts := []types.Option{
		types.WithCustomTransport(mock),
	}

	err := WithClient(context.Background(), opts, func(c *Client) error {
		return nil // function succeeds
	})

	// Since function succeeded, we currently don't return the close error
	// This test documents current behavior - may want to change this
	if err != nil {
		t.Logf("Close error is returned: %v", err)
	}

	if mock.closeCallCount != 1 {
		t.Errorf("expected Close to be called, got %d calls", mock.closeCallCount)
	}
}
