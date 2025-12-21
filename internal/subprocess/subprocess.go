package subprocess

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/victorarias/claude-agent-sdk-go/types"
)

// versionCheckTimeout is the timeout for checking CLI version.
const versionCheckTimeout = 2 * time.Second

// parseVersionOutput extracts a semver version from CLI output.
func parseVersionOutput(output string) (string, error) {
	output = strings.TrimSpace(output)
	if output == "" {
		return "", fmt.Errorf("empty version output")
	}

	// Try to find a semver pattern (X.Y.Z)
	// Handle various formats: "1.0.50", "v1.0.50", "Claude Code version 1.0.50"
	parts := strings.Fields(output)
	for _, part := range parts {
		// Remove 'v' prefix if present
		part = strings.TrimPrefix(part, "v")

		// Check if it looks like a version (contains dots and starts with digit)
		if len(part) > 0 && part[0] >= '0' && part[0] <= '9' && strings.Contains(part, ".") {
			// Validate it has at least X.Y format
			versionParts := strings.Split(part, ".")
			if len(versionParts) >= 2 {
				// Check first two parts are numeric
				if _, err := strconv.Atoi(versionParts[0]); err == nil {
					if _, err := strconv.Atoi(versionParts[1]); err == nil {
						return part, nil
					}
				}
			}
		}
	}

	// Try the whole output if it's just a version string
	output = strings.TrimPrefix(output, "v")
	if len(output) > 0 && output[0] >= '0' && output[0] <= '9' {
		versionParts := strings.Split(output, ".")
		if len(versionParts) >= 2 {
			if _, err := strconv.Atoi(versionParts[0]); err == nil {
				if _, err := strconv.Atoi(versionParts[1]); err == nil {
					return output, nil
				}
			}
		}
	}

	return "", fmt.Errorf("no version found in output: %s", output)
}

// checkCLIVersion runs the CLI to get its version.
func checkCLIVersion(cliPath string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), versionCheckTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, cliPath, "-v")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get CLI version: %w", err)
	}

	return parseVersionOutput(string(output))
}

// isVersionAtLeast checks if version a is >= version b.
func isVersionAtLeast(a, b string) bool {
	aParts := strings.Split(a, ".")
	bParts := strings.Split(b, ".")

	for i := 0; i < len(bParts) && i < len(aParts); i++ {
		aNum, aErr := strconv.Atoi(aParts[i])
		bNum, bErr := strconv.Atoi(bParts[i])

		if aErr != nil || bErr != nil {
			return false
		}

		if aNum > bNum {
			return true
		}
		if aNum < bNum {
			return false
		}
	}

	// If we get here, versions are equal up to the shorter one's length
	return len(aParts) >= len(bParts)
}

// findCLI locates the Claude CLI binary.
// Priority: explicitPath > bundledPath > PATH > common locations
func findCLI(explicitPath, bundledPath string) (string, error) {
	searchedPaths := []string{}

	// 1. Explicit path has highest priority
	if explicitPath != "" {
		searchedPaths = append(searchedPaths, explicitPath)
		if _, err := os.Stat(explicitPath); err == nil {
			return explicitPath, nil
		}
		return "", &types.CLINotFoundError{SearchedPaths: searchedPaths, CLIPath: explicitPath}
	}

	// 2. Bundled path (for packaged distributions)
	if bundledPath != "" {
		searchedPaths = append(searchedPaths, bundledPath)
		if _, err := os.Stat(bundledPath); err == nil {
			return bundledPath, nil
		}
	}

	// 3. Check PATH
	if path, err := exec.LookPath("claude"); err == nil {
		return path, nil
	}

	// 4. Check common installation locations
	home, _ := os.UserHomeDir()
	commonPaths := []string{
		// npm global (most common)
		filepath.Join(home, ".npm-global", "bin", "claude"),
		// Local bin
		"/usr/local/bin/claude",
		filepath.Join(home, ".local", "bin", "claude"),
		// Node modules
		filepath.Join(home, "node_modules", ".bin", "claude"),
		// Yarn
		filepath.Join(home, ".yarn", "bin", "claude"),
		// Claude local installation
		filepath.Join(home, ".claude", "local", "claude"),
	}

	// Windows-specific locations
	if runtime.GOOS == "windows" {
		appData := os.Getenv("APPDATA")
		localAppData := os.Getenv("LOCALAPPDATA")
		commonPaths = append(commonPaths,
			filepath.Join(appData, "npm", "claude.cmd"),
			filepath.Join(localAppData, "Programs", "claude", "claude.exe"),
			filepath.Join(home, "AppData", "Roaming", "npm", "claude.cmd"),
		)
	}

	for _, path := range commonPaths {
		searchedPaths = append(searchedPaths, path)
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", &types.CLINotFoundError{SearchedPaths: searchedPaths}
}

// WindowsMaxCommandLength is the maximum command line length on Windows.
const WindowsMaxCommandLength = 8191

// buildSettingsValue merges sandbox settings into settings JSON.
// Matches Python SDK behavior in subprocess_cli.py:118-170.
func buildSettingsValue(opts *types.Options) (string, error) {
	hasSettings := opts.Settings != ""
	hasSandbox := opts.Sandbox != nil

	if !hasSettings && !hasSandbox {
		return "", nil
	}

	// If only settings and no sandbox, pass through as-is
	if hasSettings && !hasSandbox {
		return opts.Settings, nil
	}

	// If we have sandbox settings, we need to merge into a JSON object
	settingsObj := make(map[string]any)

	if hasSettings {
		settingsStr := strings.TrimSpace(opts.Settings)
		// Check if settings is a JSON string or a file path
		if strings.HasPrefix(settingsStr, "{") && strings.HasSuffix(settingsStr, "}") {
			// Parse JSON string
			if err := json.Unmarshal([]byte(settingsStr), &settingsObj); err != nil {
				// If parsing fails, treat as file path
				// Note: File reading would go here, but for now we only support JSON strings
				return "", fmt.Errorf("failed to parse settings as JSON: %w", err)
			}
		} else {
			// It's a file path - for now, we don't support reading files during command building
			// This matches the skipped test case
			return "", fmt.Errorf("file path settings with sandbox not yet supported")
		}
	}

	// Merge sandbox settings
	if hasSandbox {
		settingsObj["sandbox"] = opts.Sandbox
	}

	data, err := json.Marshal(settingsObj)
	if err != nil {
		return "", fmt.Errorf("failed to marshal settings: %w", err)
	}

	return string(data), nil
}

// buildCommand constructs the CLI command with arguments.
func buildCommand(cliPath, prompt string, opts *types.Options, streaming bool) []string {
	cmd := []string{cliPath, "--output-format", "stream-json", "--verbose"}

	// System prompt (always include, even if empty)
	switch sp := opts.SystemPrompt.(type) {
	case string:
		if sp != "" {
			cmd = append(cmd, "--system-prompt", sp)
		} else {
			cmd = append(cmd, "--system-prompt", "")
		}
	case types.SystemPromptPreset:
		if data, err := json.Marshal(sp); err == nil {
			cmd = append(cmd, "--system-prompt", string(data))
		}
	default:
		cmd = append(cmd, "--system-prompt", "")
	}

	if opts.AppendSystemPrompt != "" {
		cmd = append(cmd, "--append-system-prompt", opts.AppendSystemPrompt)
	}

	// Tools configuration
	if len(opts.Tools) > 0 {
		cmd = append(cmd, "--tools", strings.Join(opts.Tools, ","))
	}

	if len(opts.AllowedTools) > 0 {
		cmd = append(cmd, "--allowedTools", strings.Join(opts.AllowedTools, ","))
	}

	if len(opts.DisallowedTools) > 0 {
		cmd = append(cmd, "--disallowedTools", strings.Join(opts.DisallowedTools, ","))
	}

	// Model selection
	if opts.Model != "" {
		cmd = append(cmd, "--model", opts.Model)
	}

	if opts.FallbackModel != "" {
		cmd = append(cmd, "--fallback-model", opts.FallbackModel)
	}

	// Limits
	if opts.MaxTurns > 0 {
		cmd = append(cmd, "--max-turns", strconv.Itoa(opts.MaxTurns))
	}

	if opts.MaxBudgetUSD > 0 {
		cmd = append(cmd, "--max-budget-usd", fmt.Sprintf("%.4f", opts.MaxBudgetUSD))
	}

	if opts.MaxThinkingTokens > 0 {
		cmd = append(cmd, "--max-thinking-tokens", strconv.Itoa(opts.MaxThinkingTokens))
	}

	// Permissions
	if opts.PermissionMode != "" {
		cmd = append(cmd, "--permission-mode", string(opts.PermissionMode))
	}

	if opts.PermissionPromptToolName != "" {
		cmd = append(cmd, "--permission-prompt-tool", opts.PermissionPromptToolName)
	}

	// Session management
	if opts.ContinueConversation {
		cmd = append(cmd, "--continue")
	}

	if opts.Resume != "" {
		cmd = append(cmd, "--resume", opts.Resume)
	}

	// Settings
	if opts.Settings != "" {
		cmd = append(cmd, "--settings", opts.Settings)
	}

	if len(opts.SettingSources) > 0 {
		sources := make([]string, len(opts.SettingSources))
		for i, s := range opts.SettingSources {
			sources[i] = string(s)
		}
		cmd = append(cmd, "--setting-sources", strings.Join(sources, ","))
	} else {
		// Empty setting sources to avoid loading default settings
		cmd = append(cmd, "--setting-sources", "")
	}

	// Sandbox configuration
	if opts.Sandbox != nil {
		sandboxJSON, _ := json.Marshal(opts.Sandbox)
		cmd = append(cmd, "--sandbox", string(sandboxJSON))
	}

	// Directories
	for _, dir := range opts.AddDirs {
		cmd = append(cmd, "--add-dir", dir)
	}

	// MCP servers - filter out SDK-hosted servers (type: "sdk")
	// SDK MCP servers are handled internally via control protocol, not via CLI config
	if opts.MCPServers != nil {
		filteredServers := make(map[string]any)
		switch servers := opts.MCPServers.(type) {
		case map[string]types.MCPServerConfig:
			for name, server := range servers {
				if server.Type != "sdk" {
					filteredServers[name] = server
				}
			}
		case map[string]any:
			for name, server := range servers {
				if serverMap, ok := server.(map[string]any); ok {
					if serverType, _ := serverMap["type"].(string); serverType != "sdk" {
						filteredServers[name] = server
					}
				} else {
					filteredServers[name] = server
				}
			}
		}
		if len(filteredServers) > 0 {
			config := map[string]any{"mcpServers": filteredServers}
			if data, err := json.Marshal(config); err == nil {
				cmd = append(cmd, "--mcp-config", string(data))
			}
		}
	}

	// Beta features
	if len(opts.Betas) > 0 {
		betas := make([]string, len(opts.Betas))
		for i, b := range opts.Betas {
			betas[i] = string(b)
		}
		cmd = append(cmd, "--betas", strings.Join(betas, ","))
	}

	// Streaming options
	if opts.IncludePartialMessages {
		cmd = append(cmd, "--include-partial-messages")
	}

	// Output format (JSON schema)
	if opts.OutputFormat != nil {
		if schema, ok := opts.OutputFormat["schema"]; ok {
			if data, err := json.Marshal(schema); err == nil {
				cmd = append(cmd, "--json-schema", string(data))
			}
		}
	}

	// Extra args (escape hatch for future CLI flags)
	for flag, value := range opts.ExtraArgs {
		if value == "" {
			cmd = append(cmd, "--"+flag)
		} else {
			cmd = append(cmd, "--"+flag, value)
		}
	}

	// Mode-specific args (must come last)
	if streaming {
		cmd = append(cmd, "--input-format", "stream-json")
	} else {
		cmd = append(cmd, "--print", "--", prompt)
	}

	return cmd
}

// checkCommandLength validates command length on Windows.
func checkCommandLength(cmd []string) error {
	if runtime.GOOS != "windows" {
		return nil
	}

	totalLen := 0
	for _, arg := range cmd {
		totalLen += len(arg) + 1 // +1 for space separator
	}

	if totalLen > WindowsMaxCommandLength {
		return fmt.Errorf("command length %d exceeds Windows limit of %d", totalLen, WindowsMaxCommandLength)
	}

	return nil
}

// SubprocessTransport manages the Claude CLI subprocess.
// CRITICAL: This implements write serialization for concurrent MCP tool calls.
type SubprocessTransport struct {
	prompt    string
	options   *types.Options
	streaming bool

	cliPath string
	cmd     *exec.Cmd

	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser

	messages chan map[string]any
	errors   chan error

	ready   bool
	closed  bool
	closeMu sync.Mutex

	// CRITICAL: Write mutex for concurrent MCP tool call serialization
	// All writes to stdin MUST be protected by this mutex
	writeMu sync.Mutex

	// Exit error tracking for proper error reporting
	exitError error
	exitMu    sync.Mutex

	// Temp files to clean up on close
	tempFiles []string
	tempMu    sync.Mutex

	// Stderr callback for debugging
	stderrCallback func(string)

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewSubprocessTransport creates a new subprocess transport.
func NewSubprocessTransport(prompt string, opts *types.Options) *SubprocessTransport {
	if opts == nil {
		opts = types.DefaultOptions()
	}

	return &SubprocessTransport{
		prompt:    prompt,
		options:   opts,
		streaming: prompt == "", // Empty prompt = streaming mode
		messages:  make(chan map[string]any, 100),
		errors:    make(chan error, 1),
		tempFiles: make([]string, 0),
	}
}

// NewStreamingTransport creates a transport for streaming mode.
func NewStreamingTransport(opts *types.Options) *SubprocessTransport {
	return NewSubprocessTransport("", opts)
}

// IsReady returns true if the transport is connected and not closed.
func (t *SubprocessTransport) IsReady() bool {
	t.closeMu.Lock()
	defer t.closeMu.Unlock()
	return t.ready && !t.closed
}

// Messages returns the channel of messages from the CLI.
func (t *SubprocessTransport) Messages() <-chan map[string]any {
	return t.messages
}

// Errors returns the channel of errors from the CLI.
func (t *SubprocessTransport) Errors() <-chan error {
	return t.errors
}

// SetStderrCallback sets a callback for stderr output.
func (t *SubprocessTransport) SetStderrCallback(callback func(string)) {
	t.stderrCallback = callback
}

// AddTempFile adds a temp file to be cleaned up on close.
func (t *SubprocessTransport) AddTempFile(path string) {
	t.tempMu.Lock()
	defer t.tempMu.Unlock()
	t.tempFiles = append(t.tempFiles, path)
}

// ExitError returns the exit error if the process has exited.
func (t *SubprocessTransport) ExitError() error {
	t.exitMu.Lock()
	defer t.exitMu.Unlock()
	return t.exitError
}

// Connect starts the CLI subprocess.
func (t *SubprocessTransport) Connect(ctx context.Context) error {
	t.closeMu.Lock()
	defer t.closeMu.Unlock()

	// Already connected
	if t.ready {
		return nil
	}

	// Find CLI
	cliPath, err := findCLI(t.options.CLIPath, t.options.BundledCLIPath)
	if err != nil {
		return err
	}
	t.cliPath = cliPath

	// Check CLI version unless skipped via environment variable
	if os.Getenv("CLAUDE_AGENT_SDK_SKIP_VERSION_CHECK") == "" {
		if version, err := checkCLIVersion(cliPath); err == nil {
			// Check if version meets minimum requirements
			if !isVersionAtLeast(version, types.MinimumCLIVersion) {
				return &types.CLIVersionError{
					InstalledVersion: version,
					MinimumVersion:   types.MinimumCLIVersion,
				}
			}
		}
		// If version check fails (e.g., timeout), log warning but continue
		// matching Python SDK behavior
	}

	// Build command
	args := buildCommand(cliPath, t.prompt, t.options, t.streaming)

	// Check command length on Windows
	if err := checkCommandLength(args); err != nil {
		return &types.ConnectionError{Message: err.Error()}
	}

	// Create context for cancellation
	t.ctx, t.cancel = context.WithCancel(ctx)

	// Create command
	t.cmd = exec.CommandContext(t.ctx, args[0], args[1:]...)

	// Set working directory
	if t.options.Cwd != "" {
		t.cmd.Dir = t.options.Cwd
	}

	// Set environment
	t.cmd.Env = buildEnvironment(t.options)

	// Setup pipes
	t.stdin, err = t.cmd.StdinPipe()
	if err != nil {
		return &types.ConnectionError{Message: "failed to create stdin pipe", Cause: err}
	}

	t.stdout, err = t.cmd.StdoutPipe()
	if err != nil {
		return &types.ConnectionError{Message: "failed to create stdout pipe", Cause: err}
	}

	t.stderr, err = t.cmd.StderrPipe()
	if err != nil {
		return &types.ConnectionError{Message: "failed to create stderr pipe", Cause: err}
	}

	// Start process
	if err := t.cmd.Start(); err != nil {
		return &types.ConnectionError{Message: "failed to start CLI", Cause: err}
	}

	// Start reading stdout (to be implemented in Task 5)
	t.wg.Add(1)
	go t.readMessages()

	// Start reading stderr
	t.wg.Add(1)
	go t.readStderr()

	// For non-streaming mode, close stdin immediately after start
	if !t.streaming {
		t.stdin.Close()
	}

	t.ready = true
	return nil
}

// buildEnvironment creates the environment for the subprocess.
func buildEnvironment(opts *types.Options) []string {
	env := os.Environ()

	// Add SDK-specific vars
	env = append(env, "CLAUDE_CODE_ENTRYPOINT=sdk-go")
	env = append(env, "CLAUDE_AGENT_SDK_VERSION="+types.Version)

	// Add user-provided vars
	for k, v := range opts.Env {
		env = append(env, k+"="+v)
	}

	// Add feature flags
	if opts.EnableFileCheckpointing {
		env = append(env, "CLAUDE_CODE_ENABLE_SDK_FILE_CHECKPOINTING=true")
	}

	return env
}

// parseJSONLine parses a single JSON line.
func parseJSONLine(line string) (map[string]any, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil, fmt.Errorf("empty line")
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(line), &result); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	return result, nil
}

// jsonAccumulator handles speculative JSON parsing for partial lines.
type jsonAccumulator struct {
	buffer strings.Builder
	limit  int
}

func newJSONAccumulator() *jsonAccumulator {
	return &jsonAccumulator{limit: maxBufferSize}
}

func newJSONAccumulatorWithLimit(limit int) *jsonAccumulator {
	return &jsonAccumulator{limit: limit}
}

// addLine adds a line to the accumulator and attempts to parse.
// Returns (result, nil) if JSON is complete, (nil, nil) if still accumulating,
// or (nil, error) if buffer limit is exceeded.
func (a *jsonAccumulator) addLine(line string) (map[string]any, error) {
	// Check if adding this line would exceed the buffer limit
	if a.buffer.Len()+len(line) > a.limit {
		a.buffer.Reset()
		return nil, fmt.Errorf("buffer size %d exceeds limit of %d bytes", a.buffer.Len()+len(line), a.limit)
	}

	a.buffer.WriteString(line)

	// Try to parse speculatively
	result, err := parseJSONLine(a.buffer.String())
	if err != nil {
		// Still accumulating - not an error yet
		return nil, nil
	}

	// Successfully parsed - reset buffer
	a.buffer.Reset()
	return result, nil
}

// reset clears the accumulator.
func (a *jsonAccumulator) reset() {
	a.buffer.Reset()
}

// len returns the current buffer length.
func (a *jsonAccumulator) len() int {
	return a.buffer.Len()
}

const maxBufferSize = 1024 * 1024 // 1MB

// readMessages reads JSON messages from stdout with speculative parsing.
func (t *SubprocessTransport) readMessages() {
	defer t.wg.Done()
	defer close(t.messages)

	scanner := bufio.NewScanner(t.stdout)
	buf := make([]byte, maxBufferSize)
	scanner.Buffer(buf, maxBufferSize)

	accumulator := newJSONAccumulator()

	for scanner.Scan() {
		select {
		case <-t.ctx.Done():
			return
		default:
		}

		line := scanner.Text()
		if line == "" {
			continue
		}

		// Speculative parsing - try to parse immediately
		msg, err := accumulator.addLine(line)
		if err != nil {
			// Buffer overflow - send error to error channel
			select {
			case t.errors <- err:
			default:
				// Error channel full, continue anyway
			}
			continue
		}
		if msg == nil {
			// Still accumulating, continue
			continue
		}

		// Successfully parsed
		select {
		case t.messages <- msg:
		case <-t.ctx.Done():
			return
		}
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		select {
		case t.errors <- err:
		default:
		}
	}

	// Note: cmd.Wait() is called in Close() to avoid duplicate calls
}

// readStderr reads stderr and optionally invokes callback.
func (t *SubprocessTransport) readStderr() {
	defer t.wg.Done()

	scanner := bufio.NewScanner(t.stderr)
	for scanner.Scan() {
		line := scanner.Text()
		if t.stderrCallback != nil {
			t.stderrCallback(line)
		}
	}
}

const gracefulShutdownTimeout = 5 * time.Second

// Close terminates the subprocess and cleans up resources.
// CRITICAL: Uses TOCTOU-safe pattern - state changes inside lock.
func (t *SubprocessTransport) Close() error {
	// CRITICAL: Acquire write lock FIRST to prevent concurrent writes during close
	t.writeMu.Lock()
	defer t.writeMu.Unlock()

	t.closeMu.Lock()
	if t.closed {
		t.closeMu.Unlock()
		return nil
	}
	// CRITICAL: Set closed and ready inside the lock to prevent TOCTOU
	t.closed = true
	t.ready = false
	cancel := t.cancel
	cmd := t.cmd
	stdin := t.stdin
	stdout := t.stdout
	stderr := t.stderr
	t.closeMu.Unlock()

	// Cancel context to stop goroutines
	if cancel != nil {
		cancel()
	}

	// Close stdin first to signal EOF
	if stdin != nil {
		stdin.Close()
	}

	// Terminate process if running
	if cmd != nil && cmd.Process != nil {
		done := make(chan error, 1)
		go func() {
			done <- cmd.Wait()
		}()

		select {
		case <-done:
			// Process exited gracefully
		case <-time.After(gracefulShutdownTimeout):
			// Force kill
			cmd.Process.Kill()
			<-done
		}
	}

	// Close pipes
	if stdout != nil {
		stdout.Close()
	}
	if stderr != nil {
		stderr.Close()
	}

	// Wait for goroutines
	t.wg.Wait()

	// Clean up temp files
	t.tempMu.Lock()
	for _, path := range t.tempFiles {
		os.Remove(path)
	}
	t.tempFiles = nil
	t.tempMu.Unlock()

	return nil
}

// Kill forcefully terminates the subprocess.
func (t *SubprocessTransport) Kill() error {
	t.closeMu.Lock()
	defer t.closeMu.Unlock()

	if t.cmd != nil && t.cmd.Process != nil {
		return t.cmd.Process.Kill()
	}
	return nil
}

// Write sends data to the CLI stdin with TOCTOU-safe serialization.
// CRITICAL: This method serializes concurrent writes from MCP tool calls.
func (t *SubprocessTransport) Write(data string) error {
	// CRITICAL: Acquire write lock FIRST to serialize concurrent MCP tool calls
	t.writeMu.Lock()
	defer t.writeMu.Unlock()

	// TOCTOU-safe: Check ready state INSIDE the lock
	t.closeMu.Lock()
	ready := t.ready
	closed := t.closed
	stdin := t.stdin
	t.closeMu.Unlock()

	if !ready || closed {
		return &types.ConnectionError{Message: "transport not ready for writing"}
	}
	if stdin == nil {
		return &types.ConnectionError{Message: "stdin is nil"}
	}

	// Ensure data ends with newline
	if !strings.HasSuffix(data, "\n") {
		data += "\n"
	}

	_, err := io.WriteString(stdin, data)
	if err != nil {
		return &types.ConnectionError{Message: "failed to write to stdin", Cause: err}
	}

	return nil
}

// WriteJSON marshals and writes a JSON object.
func (t *SubprocessTransport) WriteJSON(obj any) error {
	data, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return t.Write(string(data))
}

// EndInput closes stdin to signal end of input.
func (t *SubprocessTransport) EndInput() error {
	t.writeMu.Lock()
	defer t.writeMu.Unlock()

	t.closeMu.Lock()
	stdin := t.stdin
	t.closeMu.Unlock()

	if stdin != nil {
		return stdin.Close()
	}
	return nil
}
