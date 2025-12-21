package sdk

import (
	"context"
	"testing"
	"time"
)

func TestNewQuery(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	if query == nil {
		t.Fatal("NewQuery returned nil")
	}

	if query.transport != transport {
		t.Error("transport not set correctly")
	}

	if !query.streaming {
		t.Error("streaming flag not set")
	}
}

func TestQuery_Start(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	ctx := context.Background()
	err := query.Start(ctx)
	if err != nil {
		t.Errorf("Start failed: %v", err)
	}

	// Should be able to receive messages
	transport.SendMessage(map[string]any{"type": "system", "subtype": "init"})

	select {
	case msg := <-query.Messages():
		if msg.MessageType() != "system" {
			t.Errorf("got type %v, want system", msg.MessageType())
		}
	case <-time.After(time.Second):
		t.Error("timeout waiting for message")
	}

	query.Close()
}

func TestQuery_RawMessages(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer query.Close()

	transport.SendMessage(map[string]any{"type": "test", "custom": true})

	select {
	case msg := <-query.RawMessages():
		if msg["type"] != "test" {
			t.Errorf("got type %v, want test", msg["type"])
		}
	case <-time.After(time.Second):
		t.Error("timeout waiting for raw message")
	}
}

func TestQuery_ResultReceived(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer query.Close()

	// Send a result message
	transport.messages <- map[string]any{
		"type":       "result",
		"subtype":    "success",
		"session_id": "test_123",
	}

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	if !query.ResultReceived() {
		t.Error("expected result to be received")
	}
}

func TestQuery_Close(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}

	if err := query.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// Should be able to close again without error
	if err := query.Close(); err != nil {
		t.Errorf("second Close failed: %v", err)
	}
}
