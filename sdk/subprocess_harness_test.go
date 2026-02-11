// Copyright (C) 2025 Claude Agent SDK Go Contributors
// SPDX-License-Identifier: GPL-3.0-only

package sdk

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/victorarias/claude-agent-sdk-go/types"
)

func writeFakeCLIScript(t *testing.T, path string) {
	t.Helper()
	script := `#!/bin/sh
if [ "$1" = "-v" ]; then
  echo "2.0.0"
  exit 0
fi

log_file="${LOG_FILE:-/tmp/claude-sdk-go-fake-cli.log}"
while IFS= read -r line; do
  printf '%s\n' "$line" >> "$log_file"
  case "$line" in
    *"\"subtype\":\"initialize\""*)
      req_id=$(printf '%s\n' "$line" | sed -n 's/.*"request_id":"\([^"]*\)".*/\1/p')
      printf '{"type":"control_response","response":{"subtype":"success","request_id":"%s","response":{"session_id":"sess_e2e","models":[{"value":"claude-sonnet-4-5","displayName":"Sonnet 4.5","description":"test model"}],"commands":[{"name":"help","description":"Help command"}],"account":{"email":"ci@example.com"}}}}\n' "$req_id"
      ;;
    *"\"subtype\":\"set_model\""*)
      req_id=$(printf '%s\n' "$line" | sed -n 's/.*"request_id":"\([^"]*\)".*/\1/p')
      printf '{"type":"control_response","response":{"subtype":"success","request_id":"%s","response":{}}}\n' "$req_id"
      ;;
    *"\"type\":\"user\""*)
      printf '{"type":"assistant","uuid":"assistant_e2e","session_id":"sess_e2e","message":{"model":"claude-sonnet-4-5","content":[{"type":"text","text":"hello from fake cli"}]}}\n'
      printf '{"type":"result","subtype":"success","uuid":"result_e2e","duration_ms":1,"duration_api_ms":1,"is_error":false,"num_turns":1,"session_id":"sess_e2e","result":"done"}\n'
      ;;
  esac
done
`
	if err := os.WriteFile(path, []byte(script), 0755); err != nil {
		t.Fatalf("failed to write fake CLI script: %v", err)
	}
}

func TestClient_SubprocessHarness_E2E(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell-script subprocess harness is not supported on Windows")
	}

	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "fake-cli-input.log")
	cliPath := filepath.Join(tmpDir, "claude")
	writeFakeCLIScript(t, cliPath)

	client := NewClient(
		types.WithCLIPath(cliPath),
		types.WithEnv(map[string]string{
			"LOG_FILE": logPath,
		}),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer client.Close()

	if sid := client.SessionID(); sid != "sess_e2e" {
		t.Fatalf("expected session ID sess_e2e, got %q", sid)
	}

	init, err := client.InitializationResult()
	if err != nil {
		t.Fatalf("InitializationResult failed: %v", err)
	}
	if len(init.Models) != 1 || init.Models[0].Value != "claude-sonnet-4-5" {
		t.Fatalf("unexpected models from fake CLI: %+v", init.Models)
	}
	if len(init.Commands) != 1 || init.Commands[0].Name != "help" {
		t.Fatalf("unexpected commands from fake CLI: %+v", init.Commands)
	}

	if err := client.SetModel("claude-opus-4-1"); err != nil {
		t.Fatalf("SetModel failed: %v", err)
	}
	if err := client.ClearModel(); err != nil {
		t.Fatalf("ClearModel failed: %v", err)
	}

	if err := client.SendQuery("hello"); err != nil {
		t.Fatalf("SendQuery failed: %v", err)
	}
	messages, err := client.ReceiveAll()
	if err != nil {
		t.Fatalf("ReceiveAll failed: %v", err)
	}
	if len(messages) == 0 {
		t.Fatal("expected messages from fake CLI")
	}

	var gotResult bool
	for _, msg := range messages {
		if result, ok := msg.(*types.ResultMessage); ok {
			gotResult = true
			if result.Subtype != "success" {
				t.Fatalf("expected success result subtype, got %q", result.Subtype)
			}
			if result.SessionID != "sess_e2e" {
				t.Fatalf("expected sess_e2e in result, got %q", result.SessionID)
			}
		}
	}
	if !gotResult {
		t.Fatal("expected at least one result message")
	}

	// Validate control request semantics through real subprocess transport.
	file, err := os.Open(logPath)
	if err != nil {
		t.Fatalf("failed to open fake CLI log: %v", err)
	}
	defer file.Close()

	var setModelRequests []map[string]any
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, `"subtype":"set_model"`) {
			continue
		}
		var requestEnvelope map[string]any
		if err := json.Unmarshal([]byte(line), &requestEnvelope); err != nil {
			t.Fatalf("failed to parse logged set_model request: %v", err)
		}
		request, _ := requestEnvelope["request"].(map[string]any)
		setModelRequests = append(setModelRequests, request)
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("failed reading fake CLI log: %v", err)
	}

	if len(setModelRequests) < 2 {
		t.Fatalf("expected at least 2 set_model requests, got %d", len(setModelRequests))
	}
	if got, _ := setModelRequests[0]["model"].(string); got != "claude-opus-4-1" {
		t.Fatalf("expected explicit model on first request, got %+v", setModelRequests[0]["model"])
	}
	if _, ok := setModelRequests[1]["model"]; ok {
		t.Fatalf("expected clear-model request to omit model field, got %+v", setModelRequests[1]["model"])
	}
}
