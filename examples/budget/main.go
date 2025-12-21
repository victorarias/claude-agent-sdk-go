// Example: Budget control with Claude Agent SDK.
//
// This demonstrates how to use budget limits to control costs:
// - Set max_budget_usd to limit spending
// - Track accumulated costs across turns
// - Handle budget exceeded scenarios
//
// Usage:
//
//	go run examples/budget/main.go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/victorarias/claude-agent-sdk-go/sdk"
	"github.com/victorarias/claude-agent-sdk-go/types"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	fmt.Println("Budget Control Example")
	fmt.Println("======================")
	fmt.Println()

	// Example 1: Set a budget for a query
	fmt.Println("1. Single query with budget limit")
	singleQueryBudget(ctx)

	// Example 2: Track costs across multiple queries
	fmt.Println("\n2. Cost tracking across queries")
	costTrackingExample(ctx)

	// Example 3: Streaming with budget
	fmt.Println("\n3. Streaming session with budget")
	streamingBudgetExample(ctx)
}

func singleQueryBudget(ctx context.Context) {
	// Set a budget - query will be limited to this amount
	messages, err := sdk.RunQuery(ctx, "Say 'hi' briefly",
		types.WithMaxBudget(0.01), // $0.01 budget
	)
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
		return
	}

	for _, msg := range messages {
		switch m := msg.(type) {
		case *types.AssistantMessage:
			fmt.Printf("   Response: %s\n", m.Text())
		case *types.ResultMessage:
			if m.TotalCostUSD != nil {
				fmt.Printf("   Cost: $%.6f\n", *m.TotalCostUSD)
			}
		}
	}
}

func costTrackingExample(ctx context.Context) {
	var totalCost float64
	maxBudget := 0.05 // $0.05 total budget

	queries := []string{
		"Say 'one'",
		"Say 'two'",
		"Say 'three'",
	}

	for i, query := range queries {
		// Check if we have budget remaining
		remaining := maxBudget - totalCost
		if remaining <= 0 {
			fmt.Printf("   Query %d: Skipped - budget exhausted\n", i+1)
			continue
		}

		fmt.Printf("   Query %d: %s (budget remaining: $%.4f)\n", i+1, query, remaining)

		messages, err := sdk.RunQuery(ctx, query,
			types.WithMaxBudget(remaining),
		)
		if err != nil {
			fmt.Printf("   Error: %v\n", err)
			break
		}

		for _, msg := range messages {
			if result, ok := msg.(*types.ResultMessage); ok {
				if result.TotalCostUSD != nil {
					queryCost := *result.TotalCostUSD
					totalCost += queryCost
					fmt.Printf("   -> Cost: $%.6f (total: $%.6f)\n", queryCost, totalCost)
				}
			}
		}
	}

	fmt.Printf("   Final total cost: $%.6f\n", totalCost)
}

func streamingBudgetExample(ctx context.Context) {
	budget := 0.02 // $0.02 budget

	client := sdk.NewClient(
		types.WithMaxBudget(budget),
	)

	if err := client.Connect(ctx); err != nil {
		fmt.Printf("   Error: %v\n", err)
		return
	}
	defer client.Close()

	if err := client.SendQuery("What is 1+1? Be brief."); err != nil {
		fmt.Printf("   Error: %v\n", err)
		return
	}

	for {
		msg, err := client.ReceiveMessage()
		if err != nil {
			fmt.Printf("   Error: %v\n", err)
			break
		}

		switch m := msg.(type) {
		case *types.AssistantMessage:
			fmt.Printf("   Response: %s\n", m.Text())
		case *types.ResultMessage:
			if m.TotalCostUSD != nil {
				fmt.Printf("   Total cost: $%.6f (budget: $%.4f)\n", *m.TotalCostUSD, budget)
			}
			return
		}
	}
}
