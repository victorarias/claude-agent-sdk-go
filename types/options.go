package types

import (
	"fmt"
	"sync"
)

// PermissionMode controls how tool permissions are handled.
type PermissionMode string

const (
	PermissionDefault PermissionMode = "default"
	PermissionAccept  PermissionMode = "acceptEdits"
	PermissionPlan    PermissionMode = "plan"
	PermissionBypass  PermissionMode = "bypassPermissions"
)

// SettingSource specifies where settings come from.
type SettingSource string

const (
	SettingSourceUser    SettingSource = "user"
	SettingSourceProject SettingSource = "project"
	SettingSourceLocal   SettingSource = "local"
)

// SdkBeta represents beta feature flags.
type SdkBeta string

const (
	BetaContext1M SdkBeta = "context-1m-2025-08-07"
)

// AgentModel specifies the model for custom agents.
type AgentModel string

const (
	AgentModelSonnet  AgentModel = "sonnet"
	AgentModelOpus    AgentModel = "opus"
	AgentModelHaiku   AgentModel = "haiku"
	AgentModelInherit AgentModel = "inherit"
)

// AgentDefinition defines a custom agent.
type AgentDefinition struct {
	Description string     `json:"description"`
	Prompt      string     `json:"prompt"`
	Tools       []string   `json:"tools,omitempty"`
	Model       AgentModel `json:"model,omitempty"`
}

// PluginConfig defines a plugin configuration.
type PluginConfig struct {
	Type string `json:"type"` // "local"
	Path string `json:"path"`
}

// SandboxNetworkConfig defines network isolation settings.
type SandboxNetworkConfig struct {
	AllowUnixSockets    []string `json:"allowUnixSockets,omitempty"`
	AllowAllUnixSockets bool     `json:"allowAllUnixSockets,omitempty"`
	AllowLocalBinding   bool     `json:"allowLocalBinding,omitempty"`
	HTTPProxyPort       *int     `json:"httpProxyPort,omitempty"`
	SocksProxyPort      *int     `json:"socksProxyPort,omitempty"`
}

// SandboxIgnoreViolations defines violation ignore rules.
type SandboxIgnoreViolations struct {
	File    []string `json:"file,omitempty"`
	Network []string `json:"network,omitempty"`
}

// SandboxSettings defines sandbox configuration for isolation.
type SandboxSettings struct {
	Enabled                   bool                     `json:"enabled,omitempty"`
	AutoAllowBashIfSandboxed  bool                     `json:"autoAllowBashIfSandboxed,omitempty"`
	ExcludedCommands          []string                 `json:"excludedCommands,omitempty"`
	AllowUnsandboxedCommands  bool                     `json:"allowUnsandboxedCommands,omitempty"`
	Network                   *SandboxNetworkConfig    `json:"network,omitempty"`
	IgnoreViolations          *SandboxIgnoreViolations `json:"ignoreViolations,omitempty"`
	EnableWeakerNestedSandbox bool                     `json:"enableWeakerNestedSandbox,omitempty"`
}

// SystemPromptPreset allows using preset system prompts.
type SystemPromptPreset struct {
	Type   string  `json:"type"`   // "preset"
	Preset string  `json:"preset"` // "claude_code"
	Append *string `json:"append,omitempty"`
}

// ToolsPreset allows using preset tool configurations.
type ToolsPreset struct {
	Type   string `json:"type"`   // "preset"
	Preset string `json:"preset"` // "claude_code"
}

// MCPServerConfig defines an external MCP server.
type MCPServerConfig struct {
	Type    string            `json:"type"` // "stdio", "sse", "http", "sdk"
	Command string            `json:"command,omitempty"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	URL     string            `json:"url,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

// MCPServer represents an in-process SDK MCP server.
type MCPServer struct {
	Name    string
	Version string
	Tools   []*MCPTool
}

// MCPTool represents a tool in an MCP server.
type MCPTool struct {
	Name        string
	Description string
	Schema      map[string]any
	Handler     MCPToolHandler
}

// MCPToolHandler is the function signature for MCP tool handlers.
type MCPToolHandler func(args map[string]any) (*MCPToolResult, error)

// MCPToolResult is the result of an MCP tool invocation.
type MCPToolResult struct {
	Content []MCPContent `json:"content"`
	IsError bool         `json:"isError,omitempty"` // Indicates the tool execution resulted in an error
}

// MCPContent represents content in an MCP result.
type MCPContent struct {
	Type     string `json:"type"`               // "text", "image", etc.
	Text     string `json:"text,omitempty"`     // For text content
	Data     string `json:"data,omitempty"`     // For image content (base64 encoded)
	MimeType string `json:"mimeType,omitempty"` // For image content (e.g., "image/png")
}

// Version is the SDK version.
const Version = "0.1.0"

// MinimumCLIVersion is the minimum supported CLI version.
const MinimumCLIVersion = "2.0.0"

// mcpToolsByName caches tool lookups by name
var (
	mcpToolsByName   = make(map[*MCPServer]map[string]*MCPTool)
	mcpToolsByNameMu sync.RWMutex
)

// getToolsMap returns or creates the tools map for a server.
func (s *MCPServer) getToolsMap() map[string]*MCPTool {
	mcpToolsByNameMu.RLock()
	if m, ok := mcpToolsByName[s]; ok {
		mcpToolsByNameMu.RUnlock()
		return m
	}
	mcpToolsByNameMu.RUnlock()

	mcpToolsByNameMu.Lock()
	defer mcpToolsByNameMu.Unlock()

	// Double-check after acquiring write lock
	if m, ok := mcpToolsByName[s]; ok {
		return m
	}

	m := make(map[string]*MCPTool)
	for _, tool := range s.Tools {
		m[tool.Name] = tool
	}
	mcpToolsByName[s] = m
	return m
}

// GetTool returns a tool by name.
func (s *MCPServer) GetTool(name string) (*MCPTool, bool) {
	m := s.getToolsMap()
	tool, ok := m[name]
	return tool, ok
}

// CallTool calls a tool by name with the given input.
func (s *MCPServer) CallTool(name string, input map[string]any) (*MCPToolResult, error) {
	tool, ok := s.GetTool(name)
	if !ok {
		return nil, fmt.Errorf("tool not found: %s", name)
	}

	return tool.Handler(input)
}

// ToConfig returns the MCP server configuration for the CLI.
func (s *MCPServer) ToConfig() map[string]any {
	tools := make([]map[string]any, len(s.Tools))
	for i, tool := range s.Tools {
		tools[i] = map[string]any{
			"name":        tool.Name,
			"description": tool.Description,
			"inputSchema": tool.Schema,
		}
	}

	return map[string]any{
		"name":    s.Name,
		"version": s.Version,
		"tools":   tools,
	}
}

// Options configures the Claude SDK client.
type Options struct {
	// Tools configuration
	Tools           []string `json:"tools,omitempty"`
	AllowedTools    []string `json:"allowed_tools,omitempty"`
	DisallowedTools []string `json:"disallowed_tools,omitempty"`

	// System prompt - can be string or SystemPromptPreset
	SystemPrompt       any    `json:"system_prompt,omitempty"`
	AppendSystemPrompt string `json:"append_system_prompt,omitempty"`

	// Model configuration
	Model         string `json:"model,omitempty"`
	FallbackModel string `json:"fallback_model,omitempty"`

	// Permission settings
	PermissionMode           PermissionMode `json:"permission_mode,omitempty"`
	PermissionPromptToolName string         `json:"permission_prompt_tool_name,omitempty"`

	// Session settings
	ContinueConversation bool   `json:"continue_conversation,omitempty"`
	Resume               string `json:"resume,omitempty"`
	ForkSession          bool   `json:"fork_session,omitempty"`

	// Limits
	MaxTurns          int     `json:"max_turns,omitempty"`
	MaxBudgetUSD      float64 `json:"max_budget_usd,omitempty"`
	MaxThinkingTokens int     `json:"max_thinking_tokens,omitempty"`
	MaxBufferSize     int     `json:"max_buffer_size,omitempty"`

	// Paths
	Cwd           string   `json:"cwd,omitempty"`
	CLIPath       string   `json:"cli_path,omitempty"`
	BundledCLIPath string  `json:"bundled_cli_path,omitempty"`
	AddDirs       []string `json:"add_dirs,omitempty"`

	// Environment
	Env map[string]string `json:"env,omitempty"`

	// MCP Servers
	MCPServers    any                   `json:"mcp_servers,omitempty"`
	SDKMCPServers map[string]*MCPServer `json:"-"`

	// Hooks for client
	Hooks      map[HookEvent][]HookMatcher `json:"-"`
	CanUseTool CanUseToolCallback          `json:"-"`

	// Streaming
	IncludePartialMessages bool `json:"include_partial_messages,omitempty"`

	// File checkpointing
	EnableFileCheckpointing bool `json:"enable_file_checkpointing,omitempty"`

	// Output format for structured outputs
	OutputFormat map[string]any `json:"output_format,omitempty"`

	// Extra CLI arguments
	ExtraArgs map[string]string `json:"extra_args,omitempty"`

	// Beta features
	Betas []SdkBeta `json:"betas,omitempty"`

	// Settings sources
	SettingSources []SettingSource `json:"setting_sources,omitempty"`
	Settings       string          `json:"settings,omitempty"`

	// User identifier
	User string `json:"user,omitempty"`

	// Custom agents
	Agents map[string]AgentDefinition `json:"agents,omitempty"`

	// Sandbox configuration
	Sandbox *SandboxSettings `json:"sandbox,omitempty"`

	// Plugins
	Plugins []PluginConfig `json:"plugins,omitempty"`

	// Stderr callback for debugging
	StderrCallback func(string) `json:"-"`

	// Internal: custom transport for testing
	customTransport Transport `json:"-"`
}

// Option is a functional option for configuring Options.
type Option func(*Options)

// DefaultOptions returns options with sensible defaults.
func DefaultOptions() *Options {
	return &Options{
		Env:           make(map[string]string),
		MaxBufferSize: 1024 * 1024, // 1MB default
	}
}

// ApplyOptions applies a list of options to an Options struct.
func ApplyOptions(opts *Options, options ...Option) {
	for _, opt := range options {
		opt(opts)
	}
}

// CustomTransport returns the custom transport if set.
func (o *Options) CustomTransport() Transport {
	return o.customTransport
}

// SetCustomTransport sets the custom transport.
func (o *Options) SetCustomTransport(t Transport) {
	o.customTransport = t
}

// WithModel sets the model to use.
func WithModel(model string) Option {
	return func(o *Options) {
		o.Model = model
	}
}

// WithCwd sets the working directory.
func WithCwd(cwd string) Option {
	return func(o *Options) {
		o.Cwd = cwd
	}
}

// WithPermissionMode sets the permission mode.
func WithPermissionMode(mode PermissionMode) Option {
	return func(o *Options) {
		o.PermissionMode = mode
	}
}

// WithEnv sets environment variables.
func WithEnv(env map[string]string) Option {
	return func(o *Options) {
		for k, v := range env {
			o.Env[k] = v
		}
	}
}

// WithSystemPrompt sets a custom system prompt.
func WithSystemPrompt(prompt string) Option {
	return func(o *Options) {
		o.SystemPrompt = prompt
	}
}

// WithSystemPromptPreset sets a system prompt preset.
func WithSystemPromptPreset(preset SystemPromptPreset) Option {
	return func(o *Options) {
		o.SystemPrompt = preset
	}
}

// WithAppendSystemPrompt appends to the system prompt.
func WithAppendSystemPrompt(prompt string) Option {
	return func(o *Options) {
		o.AppendSystemPrompt = prompt
	}
}

// WithMaxTurns limits the number of conversation turns.
func WithMaxTurns(n int) Option {
	return func(o *Options) {
		o.MaxTurns = n
	}
}

// WithMaxBudget limits spending in USD.
func WithMaxBudget(usd float64) Option {
	return func(o *Options) {
		o.MaxBudgetUSD = usd
	}
}

// WithMaxThinkingTokens sets max thinking tokens.
func WithMaxThinkingTokens(tokens int) Option {
	return func(o *Options) {
		o.MaxThinkingTokens = tokens
	}
}

// WithTools specifies which tools to enable.
func WithTools(tools ...string) Option {
	return func(o *Options) {
		o.Tools = tools
	}
}

// WithAllowedTools specifies allowed tools.
func WithAllowedTools(tools ...string) Option {
	return func(o *Options) {
		o.AllowedTools = tools
	}
}

// WithDisallowedTools specifies disallowed tools.
func WithDisallowedTools(tools ...string) Option {
	return func(o *Options) {
		o.DisallowedTools = tools
	}
}

// WithCLIPath sets a custom path to the Claude CLI.
func WithCLIPath(path string) Option {
	return func(o *Options) {
		o.CLIPath = path
	}
}

// WithResume resumes a previous session.
func WithResume(sessionID string) Option {
	return func(o *Options) {
		o.Resume = sessionID
	}
}

// WithContinue continues the last conversation.
func WithContinue() Option {
	return func(o *Options) {
		o.ContinueConversation = true
	}
}

// WithForkSession forks the session.
func WithForkSession() Option {
	return func(o *Options) {
		o.ForkSession = true
	}
}

// WithFileCheckpointing enables file checkpointing.
func WithFileCheckpointing() Option {
	return func(o *Options) {
		o.EnableFileCheckpointing = true
	}
}

// WithPartialMessages enables partial message streaming.
func WithPartialMessages() Option {
	return func(o *Options) {
		o.IncludePartialMessages = true
	}
}

// WithOutputFormat sets structured output format.
func WithOutputFormat(format map[string]any) Option {
	return func(o *Options) {
		o.OutputFormat = format
	}
}

// WithBetas enables beta features.
func WithBetas(betas ...SdkBeta) Option {
	return func(o *Options) {
		o.Betas = betas
	}
}

// WithSettingSources sets the setting sources.
func WithSettingSources(sources ...SettingSource) Option {
	return func(o *Options) {
		o.SettingSources = sources
	}
}

// WithSettings sets the settings path or JSON.
func WithSettings(settings string) Option {
	return func(o *Options) {
		o.Settings = settings
	}
}

// WithSandbox enables sandbox mode.
func WithSandbox(sandbox SandboxSettings) Option {
	return func(o *Options) {
		o.Sandbox = &sandbox
	}
}

// WithAgents sets custom agent definitions.
func WithAgents(agents map[string]AgentDefinition) Option {
	return func(o *Options) {
		o.Agents = agents
	}
}

// WithPlugins sets plugin configurations.
func WithPlugins(plugins ...PluginConfig) Option {
	return func(o *Options) {
		o.Plugins = plugins
	}
}

// WithStderrCallback sets the stderr callback.
func WithStderrCallback(callback func(string)) Option {
	return func(o *Options) {
		o.StderrCallback = callback
	}
}

// WithMaxBufferSize sets the max buffer size for stdout.
func WithMaxBufferSize(size int) Option {
	return func(o *Options) {
		o.MaxBufferSize = size
	}
}

// WithMCPServers sets external MCP server configurations.
func WithMCPServers(servers map[string]MCPServerConfig) Option {
	return func(o *Options) {
		o.MCPServers = servers
	}
}

// WithSDKMCPServer adds an in-process MCP server.
func WithSDKMCPServer(name string, server *MCPServer) Option {
	return func(o *Options) {
		if o.SDKMCPServers == nil {
			o.SDKMCPServers = make(map[string]*MCPServer)
		}
		o.SDKMCPServers[name] = server
	}
}

// WithTransport sets a custom transport (for testing).
func WithTransport(t Transport) Option {
	return func(o *Options) {
		o.customTransport = t
	}
}

// MCPServerBuilder provides a fluent API for building MCP servers.
type MCPServerBuilder struct {
	name    string
	version string
	tools   []*MCPTool
}

// NewMCPServerBuilder creates a new MCP server builder.
func NewMCPServerBuilder(name string) *MCPServerBuilder {
	return &MCPServerBuilder{
		name:    name,
		version: "1.0.0",
		tools:   make([]*MCPTool, 0),
	}
}

// WithVersion sets the server version.
func (b *MCPServerBuilder) WithVersion(version string) *MCPServerBuilder {
	b.version = version
	return b
}

// WithTool adds a tool to the server.
func (b *MCPServerBuilder) WithTool(
	name string,
	description string,
	schema map[string]any,
	handler MCPToolHandler,
) *MCPServerBuilder {
	b.tools = append(b.tools, &MCPTool{
		Name:        name,
		Description: description,
		Schema:      schema,
		Handler:     handler,
	})
	return b
}

// Build creates the MCP server.
func (b *MCPServerBuilder) Build() *MCPServer {
	return &MCPServer{
		Name:    b.name,
		Version: b.version,
		Tools:   b.tools,
	}
}
