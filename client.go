package sdk

import (
	"context"
	"sync"
)

// Client is the high-level interface for Claude Agent SDK.
type Client struct {
	options *Options

	// MCP servers hosted by this client
	mcpServers map[string]*MCPServer

	// Hooks registered for this client
	hooks map[HookEvent][]HookMatcher

	// Permission callback
	canUseTool CanUseToolCallback

	// Transport and query
	transport Transport
	query     *Query

	// Session management
	sessionID string

	// State
	connected bool
	mu        sync.Mutex
}

// NewClient creates a new SDK client.
func NewClient(opts ...Option) *Client {
	options := DefaultOptions()
	ApplyOptions(options, opts...)

	client := &Client{
		options:    options,
		mcpServers: make(map[string]*MCPServer),
		hooks:      make(map[HookEvent][]HookMatcher),
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
	if options.customTransport != nil {
		client.transport = options.customTransport
	}

	return client
}

// Options returns the client's options.
func (c *Client) Options() *Options {
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
func WithClientMCPServer(server *MCPServer) Option {
	return func(o *Options) {
		if o.SDKMCPServers == nil {
			o.SDKMCPServers = make(map[string]*MCPServer)
		}
		o.SDKMCPServers[server.Name] = server
	}
}

// WithPreToolUseHook adds a pre-tool-use hook.
func WithPreToolUseHook(matcher map[string]any, callback HookCallback) Option {
	return func(o *Options) {
		if o.Hooks == nil {
			o.Hooks = make(map[HookEvent][]HookMatcher)
		}
		o.Hooks[HookPreToolUse] = append(o.Hooks[HookPreToolUse], HookMatcher{
			Matcher: matcher,
			Hooks:   []HookCallback{callback},
		})
	}
}

// WithPostToolUseHook adds a post-tool-use hook.
func WithPostToolUseHook(matcher map[string]any, callback HookCallback) Option {
	return func(o *Options) {
		if o.Hooks == nil {
			o.Hooks = make(map[HookEvent][]HookMatcher)
		}
		o.Hooks[HookPostToolUse] = append(o.Hooks[HookPostToolUse], HookMatcher{
			Matcher: matcher,
			Hooks:   []HookCallback{callback},
		})
	}
}

// WithStopHook adds a stop hook.
func WithStopHook(matcher map[string]any, callback HookCallback) Option {
	return func(o *Options) {
		if o.Hooks == nil {
			o.Hooks = make(map[HookEvent][]HookMatcher)
		}
		o.Hooks[HookStop] = append(o.Hooks[HookStop], HookMatcher{
			Matcher: matcher,
			Hooks:   []HookCallback{callback},
		})
	}
}

// WithUserPromptSubmitHook adds a user prompt submit hook.
func WithUserPromptSubmitHook(callback HookCallback) Option {
	return func(o *Options) {
		if o.Hooks == nil {
			o.Hooks = make(map[HookEvent][]HookMatcher)
		}
		o.Hooks[HookUserPromptSubmit] = append(o.Hooks[HookUserPromptSubmit], HookMatcher{
			Matcher: nil,
			Hooks:   []HookCallback{callback},
		})
	}
}

// WithSubagentStopHook adds a subagent stop hook.
func WithSubagentStopHook(callback HookCallback) Option {
	return func(o *Options) {
		if o.Hooks == nil {
			o.Hooks = make(map[HookEvent][]HookMatcher)
		}
		o.Hooks[HookSubagentStop] = append(o.Hooks[HookSubagentStop], HookMatcher{
			Matcher: nil,
			Hooks:   []HookCallback{callback},
		})
	}
}

// WithPreCompactHook adds a pre-compact hook.
func WithPreCompactHook(callback HookCallback) Option {
	return func(o *Options) {
		if o.Hooks == nil {
			o.Hooks = make(map[HookEvent][]HookMatcher)
		}
		o.Hooks[HookPreCompact] = append(o.Hooks[HookPreCompact], HookMatcher{
			Matcher: nil,
			Hooks:   []HookCallback{callback},
		})
	}
}

// WithHookTimeout adds a hook with a timeout.
func WithHookTimeout(event HookEvent, matcher map[string]any, timeout float64, callback HookCallback) Option {
	return func(o *Options) {
		if o.Hooks == nil {
			o.Hooks = make(map[HookEvent][]HookMatcher)
		}
		o.Hooks[event] = append(o.Hooks[event], HookMatcher{
			Matcher: matcher,
			Hooks:   []HookCallback{callback},
			Timeout: &timeout,
		})
	}
}

// WithCanUseTool sets the tool permission callback.
func WithCanUseTool(callback CanUseToolCallback) Option {
	return func(o *Options) {
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
			c.transport = NewStreamingTransport(c.options)
		} else {
			c.transport = NewSubprocessTransport(prompt, c.options)
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
