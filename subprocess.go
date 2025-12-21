package sdk

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
)

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
		return "", &CLINotFoundError{SearchedPaths: searchedPaths, CLIPath: explicitPath}
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

	return "", &CLINotFoundError{SearchedPaths: searchedPaths}
}

// WindowsMaxCommandLength is the maximum command line length on Windows.
const WindowsMaxCommandLength = 8191

// buildCommand constructs the CLI command with arguments.
func buildCommand(cliPath, prompt string, opts *Options, streaming bool) []string {
	cmd := []string{cliPath, "--output-format", "stream-json", "--verbose"}

	// System prompt (always include, even if empty)
	switch sp := opts.SystemPrompt.(type) {
	case string:
		if sp != "" {
			cmd = append(cmd, "--system-prompt", sp)
		} else {
			cmd = append(cmd, "--system-prompt", "")
		}
	case SystemPromptPreset:
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
		case map[string]MCPServerConfig:
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
