// Package sdk provides a Go client for the Claude Agent SDK.
//
// The SDK spawns the Claude CLI as a subprocess and communicates
// via JSON streaming for bidirectional control protocol.
package sdk

// Version is the SDK version.
const Version = "0.1.0"

// MinimumCLIVersion is the minimum supported CLI version.
const MinimumCLIVersion = "2.0.0"
