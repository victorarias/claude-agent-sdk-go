// Copyright (C) 2025 Claude Agent SDK Go Contributors
// SPDX-License-Identifier: GPL-3.0-only

package sdk

import (
	"context"
	"fmt"

	"github.com/victorarias/claude-agent-sdk-go/types"
)

// UnstableV2Session is a lightweight multi-turn session API aligned with TS unstable_v2.
type UnstableV2Session struct {
	client *Client
}

// UnstableV2CreateSession creates and connects a new session.
func UnstableV2CreateSession(ctx context.Context, opts ...types.Option) (*UnstableV2Session, error) {
	client := NewClient(opts...)
	if err := client.Connect(ctx); err != nil {
		return nil, err
	}
	return &UnstableV2Session{client: client}, nil
}

// UnstableV2ResumeSession resumes a previously created session ID.
func UnstableV2ResumeSession(ctx context.Context, sessionID string, opts ...types.Option) (*UnstableV2Session, error) {
	opts = append(opts, types.WithResume(sessionID))
	return UnstableV2CreateSession(ctx, opts...)
}

// UnstableV2Prompt runs a one-shot prompt and returns the result message.
func UnstableV2Prompt(ctx context.Context, message string, opts ...types.Option) (*types.ResultMessage, error) {
	messages, err := RunQuery(ctx, message, opts...)
	if err != nil {
		return nil, err
	}
	for i := len(messages) - 1; i >= 0; i-- {
		if result, ok := messages[i].(*types.ResultMessage); ok {
			return result, nil
		}
	}
	return nil, fmt.Errorf("no result message received")
}

// SessionID returns the current session ID if available.
func (s *UnstableV2Session) SessionID() (string, error) {
	if s == nil || s.client == nil {
		return "", fmt.Errorf("session is not initialized")
	}
	if sid := s.client.SessionID(); sid != "" {
		return sid, nil
	}
	return "", fmt.Errorf("session ID not available yet")
}

// Send sends a user message in the active session.
func (s *UnstableV2Session) Send(message string) error {
	if s == nil || s.client == nil {
		return fmt.Errorf("session is not initialized")
	}
	return s.client.SendQuery(message)
}

// Stream returns the streaming output channel for the session.
func (s *UnstableV2Session) Stream() <-chan types.Message {
	if s == nil || s.client == nil {
		ch := make(chan types.Message)
		close(ch)
		return ch
	}
	return s.client.Messages()
}

// Close closes the underlying client and terminates the session.
func (s *UnstableV2Session) Close() error {
	if s == nil || s.client == nil {
		return nil
	}
	return s.client.Close()
}
