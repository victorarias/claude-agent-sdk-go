package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// Query handles the bidirectional control protocol.
type Query struct {
	transport Transport
	streaming bool

	// Control protocol state
	pendingRequests map[string]chan map[string]any
	pendingMu       sync.Mutex
	requestCounter  atomic.Uint64

	// Hook callbacks
	hookCallbacks  map[string]HookCallback
	nextCallbackID atomic.Uint64
	hookMu         sync.RWMutex

	// Permission callback
	canUseTool CanUseToolCallback

	// Message channels
	messages    chan Message        // Parsed messages
	rawMessages chan map[string]any // Raw messages for custom handling
	errors      chan error

	// Result tracking
	resultReceived atomic.Bool
	lastResult     *ResultMessage
	resultMu       sync.RWMutex

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
func NewQuery(transport Transport, streaming bool) *Query {
	return &Query{
		transport:       transport,
		streaming:       streaming,
		pendingRequests: make(map[string]chan map[string]any),
		hookCallbacks:   make(map[string]HookCallback),
		messages:        make(chan Message, 100),
		rawMessages:     make(chan map[string]any, 100),
		errors:          make(chan error, 1),
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
func (q *Query) Messages() <-chan Message {
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
func (q *Query) LastResult() *ResultMessage {
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
	msg, err := ParseMessage(raw)
	if err != nil {
		// Still send to channel for unknown message types
		select {
		case q.errors <- err:
		default:
		}
		return
	}

	// Track result messages
	if result, ok := msg.(*ResultMessage); ok {
		q.resultMu.Lock()
		q.lastResult = result
		q.resultMu.Unlock()
		q.resultReceived.Store(true)
	}

	// Send parsed message
	select {
	case q.messages <- msg:
	case <-q.ctx.Done():
	}
}

// handleControlResponse routes a control response to the waiting request.
func (q *Query) handleControlResponse(msg map[string]any) {
	response, ok := msg["response"].(map[string]any)
	if !ok {
		return
	}

	requestID, _ := response["request_id"].(string)
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
	if response["subtype"] == "error" {
		errMsg, _ := response["error"].(string)
		select {
		case respChan <- map[string]any{"error": errMsg}:
		default:
		}
		return
	}

	// Send success response
	respData, _ := response["response"].(map[string]any)
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
	// TODO: Implement in Task 7
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

	// Generate request ID
	id := q.requestCounter.Add(1)
	requestID := fmt.Sprintf("req_%d", id)

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
