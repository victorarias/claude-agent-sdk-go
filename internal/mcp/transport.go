package mcp

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"sync"
)

// MCPServerTransport handles stdio communication for an MCP server.
type MCPServerTransport struct {
	handler *MCPHandler
	reader  *bufio.Reader
	writer  io.Writer
	mu      sync.Mutex
}

// NewMCPServerTransport creates a new transport for the given server.
func NewMCPServerTransport(server *MCPServer, input io.Reader, output io.Writer) *MCPServerTransport {
	return &MCPServerTransport{
		handler: NewMCPHandler(server),
		reader:  bufio.NewReader(input),
		writer:  output,
	}
}

// Run processes messages until context is cancelled or input is closed.
func (t *MCPServerTransport) Run(ctx context.Context) error {
	// Use a channel to signal when a line is ready
	lineCh := make(chan []byte, 1)
	errCh := make(chan error, 1)

	go func() {
		for {
			line, err := t.reader.ReadBytes('\n')
			if err != nil {
				errCh <- err
				return
			}
			lineCh <- line
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-errCh:
			if err == io.EOF {
				return nil
			}
			return err
		case line := <-lineCh:
			if err := t.processLine(line); err != nil {
				// Log error but continue processing
				continue
			}
		}
	}
}

// ProcessOne reads and handles a single message.
func (t *MCPServerTransport) ProcessOne() error {
	line, err := t.reader.ReadBytes('\n')
	if err != nil {
		return err
	}

	return t.processLine(line)
}

// processLine processes a single line of input.
func (t *MCPServerTransport) processLine(line []byte) error {
	// Skip empty lines
	line = bytes.TrimSpace(line)
	if len(line) == 0 {
		return nil
	}

	// Handle the request
	respBytes, err := t.handler.HandleBytes(line)
	if err != nil {
		return err
	}

	// Write response if present
	if respBytes != nil {
		t.mu.Lock()
		_, err := t.writer.Write(respBytes)
		if err == nil {
			_, err = t.writer.Write([]byte("\n"))
		}
		t.mu.Unlock()
		return err
	}

	return nil
}

// Server returns the underlying MCP server.
func (t *MCPServerTransport) Server() *MCPServer {
	return t.handler.server
}
