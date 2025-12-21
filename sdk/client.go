package sdk

import (
	"context"
	"sync"

	"github.com/victorarias/claude-agent-sdk-go/internal/parser"
	"github.com/victorarias/claude-agent-sdk-go/internal/subprocess"
	"github.com/victorarias/claude-agent-sdk-go/types"
)

// Client is the high-level interface for Claude Agent SDK.
type Client struct {
	options *types.Options

	// MCP servers hosted by this client
	mcpServers map[string]*types.MCPServer

	// Hooks registered for this client
	hooks map[types.HookEvent][]types.HookMatcher

	// Permission callback
	canUseTool types.CanUseToolCallback

	// Transport and query
	transport types.Transport
	query     *Query

	// Session management
	sessionID string

	// State
	connected bool
	mu        sync.Mutex
}

// NewClient creates a new SDK client.
func NewClient(opts ...types.Option) *Client {
	options := types.DefaultOptions()
	types.ApplyOptions(options, opts...)

	client := &Client{
		options:    options,
		mcpServers: make(map[string]*types.MCPServer),
		hooks:      make(map[types.HookEvent][]types.HookMatcher),
	}

	// Copy SDK MCP servers from options
	if options.SDKMCPServers != nil {
		for name, server := range options.SDKMCPServers {
			client.mcpServers[name] = server
		}
	}

	// Copy hooks from options
	if options.Hooks != nil {
		for event, matchers := range options.Hooks {
			client.hooks[event] = matchers
		}
	}

	// Copy canUseTool from options
	if options.CanUseTool != nil {
		client.canUseTool = options.CanUseTool
	}

	// Use transport if provided in options
	if options.CustomTransport() != nil {
		client.transport = options.CustomTransport()
	}

	return client
}

// Options returns the client's options.
func (c *Client) Options() *types.Options {
	return c.options
}

// IsConnected returns true if the client is connected.
func (c *Client) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.connected
}

// SessionID returns the current session ID.
func (c *Client) SessionID() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.sessionID
}

// WithClientMCPServer adds an MCP server to the client.
func WithClientMCPServer(server *types.MCPServer) types.Option {
	return func(o *types.Options) {
		if o.SDKMCPServers == nil {
			o.SDKMCPServers = make(map[string]*types.MCPServer)
		}
		o.SDKMCPServers[server.Name] = server
	}
}

// WithPreToolUseHook adds a pre-tool-use hook.
func WithPreToolUseHook(matcher map[string]any, callback types.HookCallback) types.Option {
	return func(o *types.Options) {
		if o.Hooks == nil {
			o.Hooks = make(map[types.HookEvent][]types.HookMatcher)
		}
		o.Hooks[types.HookPreToolUse] = append(o.Hooks[types.HookPreToolUse], types.HookMatcher{
			Matcher: matcher,
			Hooks:   []types.HookCallback{callback},
		})
	}
}

// WithPostToolUseHook adds a post-tool-use hook.
func WithPostToolUseHook(matcher map[string]any, callback types.HookCallback) types.Option {
	return func(o *types.Options) {
		if o.Hooks == nil {
			o.Hooks = make(map[types.HookEvent][]types.HookMatcher)
		}
		o.Hooks[types.HookPostToolUse] = append(o.Hooks[types.HookPostToolUse], types.HookMatcher{
			Matcher: matcher,
			Hooks:   []types.HookCallback{callback},
		})
	}
}

// WithStopHook adds a stop hook.
func WithStopHook(matcher map[string]any, callback types.HookCallback) types.Option {
	return func(o *types.Options) {
		if o.Hooks == nil {
			o.Hooks = make(map[types.HookEvent][]types.HookMatcher)
		}
		o.Hooks[types.HookStop] = append(o.Hooks[types.HookStop], types.HookMatcher{
			Matcher: matcher,
			Hooks:   []types.HookCallback{callback},
		})
	}
}

// WithUserPromptSubmitHook adds a user prompt submit hook.
func WithUserPromptSubmitHook(callback types.HookCallback) types.Option {
	return func(o *types.Options) {
		if o.Hooks == nil {
			o.Hooks = make(map[types.HookEvent][]types.HookMatcher)
		}
		o.Hooks[types.HookUserPromptSubmit] = append(o.Hooks[types.HookUserPromptSubmit], types.HookMatcher{
			Matcher: nil,
			Hooks:   []types.HookCallback{callback},
		})
	}
}

// WithSubagentStopHook adds a subagent stop hook.
func WithSubagentStopHook(callback types.HookCallback) types.Option {
	return func(o *types.Options) {
		if o.Hooks == nil {
			o.Hooks = make(map[types.HookEvent][]types.HookMatcher)
		}
		o.Hooks[types.HookSubagentStop] = append(o.Hooks[types.HookSubagentStop], types.HookMatcher{
			Matcher: nil,
			Hooks:   []types.HookCallback{callback},
		})
	}
}

// WithPreCompactHook adds a pre-compact hook.
func WithPreCompactHook(callback types.HookCallback) types.Option {
	return func(o *types.Options) {
		if o.Hooks == nil {
			o.Hooks = make(map[types.HookEvent][]types.HookMatcher)
		}
		o.Hooks[types.HookPreCompact] = append(o.Hooks[types.HookPreCompact], types.HookMatcher{
			Matcher: nil,
			Hooks:   []types.HookCallback{callback},
		})
	}
}

// WithHookTimeout adds a hook with a timeout.
func WithHookTimeout(event types.HookEvent, matcher map[string]any, timeout float64, callback types.HookCallback) types.Option {
	return func(o *types.Options) {
		if o.Hooks == nil {
			o.Hooks = make(map[types.HookEvent][]types.HookMatcher)
		}
		o.Hooks[event] = append(o.Hooks[event], types.HookMatcher{
			Matcher: matcher,
			Hooks:   []types.HookCallback{callback},
			Timeout: &timeout,
		})
	}
}

// WithCanUseTool sets the tool permission callback.
func WithCanUseTool(callback types.CanUseToolCallback) types.Option {
	return func(o *types.Options) {
		o.CanUseTool = callback
	}
}

// Connect establishes a connection to Claude in streaming mode.
func (c *Client) Connect(ctx context.Context) error {
	return c.connect(ctx, "", true)
}

// ConnectWithPrompt establishes a connection with an initial prompt (non-streaming).
func (c *Client) ConnectWithPrompt(ctx context.Context, prompt string) error {
	return c.connect(ctx, prompt, false)
}

// connect is the internal connection method.
func (c *Client) connect(ctx context.Context, prompt string, streaming bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connected {
		return nil
	}

	// Create transport if not provided (for testing)
	if c.transport == nil {
		if streaming {
			c.transport = subprocess.NewStreamingTransport(c.options)
		} else {
			c.transport = subprocess.NewSubprocessTransport(prompt, c.options)
		}
	}

	// Connect transport
	if err := c.transport.Connect(ctx); err != nil {
		return err
	}

	// Create query
	c.query = NewQuery(c.transport, streaming)

	// Set permission callback
	if c.canUseTool != nil {
		c.query.SetCanUseTool(c.canUseTool)
	}

	// Register MCP servers
	for _, server := range c.mcpServers {
		c.query.RegisterMCPServer(server)
	}

	// Start query
	if err := c.query.Start(ctx); err != nil {
		c.transport.Close()
		return err
	}

	// Initialize if streaming
	if streaming {
		result, err := c.query.Initialize(c.hooks)
		if err != nil {
			c.transport.Close()
			return err
		}

		// Extract session ID from init response
		if result != nil {
			if sid, ok := result["session_id"].(string); ok {
				c.sessionID = sid
			}
		}
	}

	c.connected = true
	return nil
}

// Disconnect closes the connection.
func (c *Client) Disconnect() error {
	return c.Close()
}

// Close closes the client and releases resources.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return nil
	}

	c.connected = false

	// Unregister MCP servers
	if c.query != nil {
		for name := range c.mcpServers {
			c.query.UnregisterMCPServer(name)
		}
		c.query.Close()
		c.query = nil
	}

	if c.transport != nil {
		c.transport.Close()
		c.transport = nil
	}

	return nil
}

// RunQuery performs a one-shot query and returns all messages.
func RunQuery(ctx context.Context, prompt string, opts ...types.Option) ([]types.Message, error) {
	options := types.DefaultOptions()
	types.ApplyOptions(options, opts...)

	// Extract transport if provided
	var transport types.Transport
	if options.CustomTransport() != nil {
		transport = options.CustomTransport()
	} else {
		transport = subprocess.NewSubprocessTransport(prompt, options)
	}

	// Connect
	if err := transport.Connect(ctx); err != nil {
		return nil, err
	}
	defer transport.Close()

	// Collect messages
	var messages []types.Message

	for msg := range transport.Messages() {
		parsed, err := parser.ParseMessage(msg)
		if err != nil {
			// Skip unparseable messages
			continue
		}
		messages = append(messages, parsed)

		// Stop on result
		if _, ok := parsed.(*types.ResultMessage); ok {
			break
		}
	}

	return messages, nil
}

// QueryStream performs a query and streams messages back.
func QueryStream(ctx context.Context, prompt string, opts ...types.Option) (<-chan types.Message, <-chan error) {
	msgChan := make(chan types.Message, 100)
	errChan := make(chan error, 1)

	go func() {
		defer close(msgChan)
		defer close(errChan)

		options := types.DefaultOptions()
		types.ApplyOptions(options, opts...)

		var transport types.Transport
		if options.CustomTransport() != nil {
			transport = options.CustomTransport()
		} else {
			transport = subprocess.NewSubprocessTransport(prompt, options)
		}

		if err := transport.Connect(ctx); err != nil {
			errChan <- err
			return
		}
		defer transport.Close()

		for msg := range transport.Messages() {
			select {
			case <-ctx.Done():
				errChan <- ctx.Err()
				return
			default:
			}

			parsed, err := parser.ParseMessage(msg)
			if err != nil {
				continue
			}

			select {
			case msgChan <- parsed:
			case <-ctx.Done():
				errChan <- ctx.Err()
				return
			}

			if _, ok := parsed.(*types.ResultMessage); ok {
				return
			}
		}
	}()

	return msgChan, errChan
}

// SendQuery sends a query in streaming mode.
func (c *Client) SendQuery(prompt string, sessionID ...string) error {
	c.mu.Lock()
	if !c.connected || c.query == nil {
		c.mu.Unlock()
		return &types.ConnectionError{Message: "not connected"}
	}
	q := c.query
	sid := c.sessionID
	c.mu.Unlock()

	if len(sessionID) > 0 {
		sid = sessionID[0]
	}
	if sid == "" {
		sid = "default"
	}

	return q.SendUserMessage(prompt, sid)
}

// ReceiveMessage receives the next message.
func (c *Client) ReceiveMessage() (Message, error) {
	c.mu.Lock()
	if !c.connected || c.query == nil {
		c.mu.Unlock()
		return nil, &types.ConnectionError{Message: "not connected"}
	}
	q := c.query
	c.mu.Unlock()

	select {
	case msg, ok := <-q.Messages():
		if !ok {
			return nil, &types.ConnectionError{Message: "channel closed"}
		}
		return msg, nil
	case err := <-q.Errors():
		return nil, err
	}
}

// ReceiveAll receives all messages until result.
func (c *Client) ReceiveAll() ([]types.Message, error) {
	var messages []types.Message
	for {
		msg, err := c.ReceiveMessage()
		if err != nil {
			return messages, err
		}
		messages = append(messages, msg)
		if _, ok := msg.(*types.ResultMessage); ok {
			return messages, nil
		}
	}
}

// ReceiveResponse sends a query and receives all response messages.
func (c *Client) ReceiveResponse(prompt string) ([]types.Message, error) {
	if err := c.SendQuery(prompt); err != nil {
		return nil, err
	}
	return c.ReceiveAll()
}

// Interrupt sends an interrupt signal.
func (c *Client) Interrupt() error {
	c.mu.Lock()
	if !c.connected || c.query == nil {
		c.mu.Unlock()
		return &types.ConnectionError{Message: "not connected"}
	}
	q := c.query
	c.mu.Unlock()

	return q.Interrupt()
}

// SetPermissionMode changes the permission mode.
func (c *Client) SetPermissionMode(mode types.PermissionMode) error {
	c.mu.Lock()
	if !c.connected || c.query == nil {
		c.mu.Unlock()
		return &types.ConnectionError{Message: "not connected"}
	}
	q := c.query
	c.mu.Unlock()

	return q.SetPermissionMode(mode)
}

// SetModel changes the AI model.
func (c *Client) SetModel(model string) error {
	c.mu.Lock()
	if !c.connected || c.query == nil {
		c.mu.Unlock()
		return &types.ConnectionError{Message: "not connected"}
	}
	q := c.query
	c.mu.Unlock()

	return q.SetModel(model)
}

// RewindFiles rewinds tracked files to a specific user message.
func (c *Client) RewindFiles(userMessageID string) error {
	c.mu.Lock()
	if !c.connected || c.query == nil {
		c.mu.Unlock()
		return &types.ConnectionError{Message: "not connected"}
	}
	q := c.query
	c.mu.Unlock()

	return q.RewindFiles(userMessageID)
}

// ServerInfo returns the initialization info.
func (c *Client) ServerInfo() map[string]any {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.query != nil {
		return c.query.InitResult()
	}
	return nil
}

// ResultReceived returns true if a result has been received.
func (c *Client) ResultReceived() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.query != nil {
		return c.query.ResultReceived()
	}
	return false
}

// LastResult returns the last result message.
func (c *Client) LastResult() *types.ResultMessage {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.query != nil {
		return c.query.LastResult()
	}
	return nil
}

// ClientFunc is a function that uses a client.
type ClientFunc func(*Client) error

// WithClient creates a client, connects, runs the function, and ensures cleanup.
func WithClient(ctx context.Context, opts []Option, fn ClientFunc) error {
	client := NewClient(opts...)

	if err := client.Connect(ctx); err != nil {
		return err
	}
	defer client.Close()

	return fn(client)
}

// Run connects and runs a function with the client.
func (c *Client) Run(ctx context.Context, fn func() error) error {
	if err := c.Connect(ctx); err != nil {
		return err
	}
	defer c.Close()
	return fn()
}

// types.Messages returns a channel that yields messages until closed or error.
// Use this for iterating over responses in streaming mode.
func (c *Client) types.Messages() <-chan types.Message {
	c.mu.Lock()
	if !c.connected || c.query == nil {
		c.mu.Unlock()
		ch := make(chan types.Message)
		close(ch)
		return ch
	}
	q := c.query
	c.mu.Unlock()

	return q.Messages()
}

// RawMessages returns a channel of raw message maps.
func (c *Client) RawMessages() <-chan map[string]any {
	c.mu.Lock()
	if !c.connected || c.query == nil {
		c.mu.Unlock()
		ch := make(chan map[string]any)
		close(ch)
		return ch
	}
	q := c.query
	c.mu.Unlock()

	return q.RawMessages()
}

// Errors returns the error channel.
func (c *Client) Errors() <-chan error {
	c.mu.Lock()
	if !c.connected || c.query == nil {
		c.mu.Unlock()
		ch := make(chan error, 1)
		close(ch)
		return ch
	}
	q := c.query
	c.mu.Unlock()

	return q.Errors()
}
