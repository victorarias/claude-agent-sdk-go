// Example: SDK-hosted MCP server with custom tools.
//
// This demonstrates how to create custom tools that Claude can use.
// Tools are hosted in-process by the SDK and exposed via MCP protocol.
//
// Usage:
//
//	go run examples/mcp-server/main.go
package main

import (
	"context"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/victorarias/claude-agent-sdk-go/sdk"
	"github.com/victorarias/claude-agent-sdk-go/types"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Build custom MCP server with tools
	mathServer := types.NewMCPServerBuilder("math-tools").
		// Calculator tool
		WithTool("calculate", "Perform mathematical calculations", map[string]any{
			"type": "object",
			"properties": map[string]any{
				"a": map[string]any{
					"type":        "number",
					"description": "First number",
				},
				"b": map[string]any{
					"type":        "number",
					"description": "Second number",
				},
				"operation": map[string]any{
					"type":        "string",
					"description": "Operation to perform (add, subtract, multiply, divide, power)",
					"enum":        []string{"add", "subtract", "multiply", "divide", "power"},
				},
			},
			"required": []string{"a", "b", "operation"},
		}, func(input map[string]any) (*types.MCPToolResult, error) {
			a, _ := input["a"].(float64)
			b, _ := input["b"].(float64)
			op, _ := input["operation"].(string)

			var result float64
			switch op {
			case "add":
				result = a + b
			case "subtract":
				result = a - b
			case "multiply":
				result = a * b
			case "divide":
				if b == 0 {
					return nil, fmt.Errorf("cannot divide by zero")
				}
				result = a / b
			case "power":
				result = math.Pow(a, b)
			default:
				return nil, fmt.Errorf("unknown operation: %s", op)
			}

			return &types.MCPToolResult{
				Content: []types.MCPContent{{
					Type: "text",
					Text: fmt.Sprintf("%.4f %s %.4f = %.4f", a, op, b, result),
				}},
			}, nil
		}).
		// Unit converter tool
		WithTool("convert_units", "Convert between units", map[string]any{
			"type": "object",
			"properties": map[string]any{
				"value": map[string]any{
					"type":        "number",
					"description": "Value to convert",
				},
				"from": map[string]any{
					"type":        "string",
					"description": "Source unit (km, miles, celsius, fahrenheit, kg, lbs)",
				},
				"to": map[string]any{
					"type":        "string",
					"description": "Target unit",
				},
			},
			"required": []string{"value", "from", "to"},
		}, func(input map[string]any) (*types.MCPToolResult, error) {
			value, _ := input["value"].(float64)
			from, _ := input["from"].(string)
			to, _ := input["to"].(string)

			result, err := convertUnits(value, from, to)
			if err != nil {
				return nil, err
			}

			return &types.MCPToolResult{
				Content: []types.MCPContent{{
					Type: "text",
					Text: fmt.Sprintf("%.4f %s = %.4f %s", value, from, result, to),
				}},
			}, nil
		}).
		Build()

	// Build a data lookup server
	dataServer := types.NewMCPServerBuilder("data-lookup").
		// Timezone tool
		WithTool("get_timezone", "Get current time in a timezone", map[string]any{
			"type": "object",
			"properties": map[string]any{
				"timezone": map[string]any{
					"type":        "string",
					"description": "Timezone name (e.g., 'America/New_York', 'Europe/London', 'Asia/Tokyo')",
				},
			},
			"required": []string{"timezone"},
		}, func(input map[string]any) (*types.MCPToolResult, error) {
			tzName, _ := input["timezone"].(string)
			loc, err := time.LoadLocation(tzName)
			if err != nil {
				return nil, fmt.Errorf("unknown timezone: %s", tzName)
			}
			now := time.Now().In(loc)
			return &types.MCPToolResult{
				Content: []types.MCPContent{{
					Type: "text",
					Text: fmt.Sprintf("Time in %s: %s (%s)", tzName, now.Format("15:04:05"), now.Format("2006-01-02")),
				}},
			}, nil
		}).
		Build()

	// Create client with MCP servers
	client := sdk.NewClient(
		types.WithModel("claude-sonnet-4-5"),
		sdk.WithClientMCPServer(mathServer),
		sdk.WithClientMCPServer(dataServer),
	)

	// Connect
	if err := client.Connect(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	fmt.Println("MCP Server Example - Custom tools available")
	fmt.Println("--------------------------------------------")
	fmt.Println()
	fmt.Println("Available tools:")
	fmt.Println("  - calculate: Perform math operations")
	fmt.Println("  - convert_units: Convert between units")
	fmt.Println("  - get_timezone: Get time in any timezone")
	fmt.Println()

	// Example prompts that will use our custom tools
	prompts := []string{
		"What is 15 multiplied by 7, and then add 23?",
		"Convert 100 kilometers to miles.",
		"What time is it right now in Tokyo?",
	}

	for i, prompt := range prompts {
		fmt.Printf("--- Query %d ---\n", i+1)
		fmt.Printf("Q: %s\n\n", prompt)

		if err := client.SendQuery(prompt); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			continue
		}

		fmt.Print("A: ")
		for {
			msg, err := client.ReceiveMessage()
			if err != nil {
				fmt.Fprintf(os.Stderr, "\nError: %v\n", err)
				break
			}

			switch m := msg.(type) {
			case *types.AssistantMessage:
				fmt.Print(m.Text())

				// Show which tools were used
				if m.HasToolCalls() {
					fmt.Print("\n   [Tools used: ")
					for j, tc := range m.ToolCalls() {
						if j > 0 {
							fmt.Print(", ")
						}
						fmt.Print(tc.Name)
					}
					fmt.Print("]")
				}
			case *types.ResultMessage:
				fmt.Println()
				goto nextQuery
			}
		}
	nextQuery:
		fmt.Println()
	}
}

// Unit converter
func convertUnits(value float64, from, to string) (float64, error) {
	conversions := map[string]map[string]func(float64) float64{
		"km": {
			"miles": func(v float64) float64 { return v * 0.621371 },
		},
		"miles": {
			"km": func(v float64) float64 { return v * 1.60934 },
		},
		"celsius": {
			"fahrenheit": func(v float64) float64 { return v*9/5 + 32 },
		},
		"fahrenheit": {
			"celsius": func(v float64) float64 { return (v - 32) * 5 / 9 },
		},
		"kg": {
			"lbs": func(v float64) float64 { return v * 2.20462 },
		},
		"lbs": {
			"kg": func(v float64) float64 { return v * 0.453592 },
		},
	}

	if fromMap, ok := conversions[from]; ok {
		if fn, ok := fromMap[to]; ok {
			return math.Round(fn(value)*10000) / 10000, nil
		}
	}
	return 0, fmt.Errorf("cannot convert from %s to %s", from, to)
}
