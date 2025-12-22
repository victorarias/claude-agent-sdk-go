// Example: System prompt variations for different use cases.
//
// This example demonstrates various system prompt configuration patterns:
// - Domain-specific assistant configurations (technical writer, educator, analyst)
// - Combining presets with custom instructions
// - Different prompt styles for different use cases
// - Multi-faceted prompts with specific guidelines
//
// Usage:
//
//	go run examples/prompt-variations/main.go [example-number]
//
// Examples:
//   1. Technical Writer - Clear, structured documentation style
//   2. Code Educator - Patient, educational coding explanations
//   3. Security Analyst - Security-focused code review
//   4. API Designer - REST API design with best practices
//   5. Performance Optimizer - Performance-focused analysis
//   6. Preset + Custom - Combining claude_code with domain context
//   7. Tone & Style - Different communication styles
//   8. Multi-Language - Bilingual or multilingual responses
package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/victorarias/claude-agent-sdk-go/sdk"
	"github.com/victorarias/claude-agent-sdk-go/types"
)

func main() {
	// Parse example number from command line
	exampleNum := 0
	if len(os.Args) > 1 {
		if num, err := strconv.Atoi(os.Args[1]); err == nil {
			exampleNum = num
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	fmt.Println("System Prompt Variations")
	fmt.Println("========================")
	fmt.Println()

	// Run specific example or all examples
	examples := []struct {
		name string
		fn   func(context.Context)
	}{
		{"Technical Writer", technicalWriterExample},
		{"Code Educator", codeEducatorExample},
		{"Security Analyst", securityAnalystExample},
		{"API Designer", apiDesignerExample},
		{"Performance Optimizer", performanceOptimizerExample},
		{"Preset + Custom Instructions", presetWithCustomExample},
		{"Tone & Style Variations", toneAndStyleExample},
		{"Multi-Language Support", multiLanguageExample},
	}

	if exampleNum > 0 && exampleNum <= len(examples) {
		// Run specific example
		example := examples[exampleNum-1]
		fmt.Printf("Running Example %d: %s\n", exampleNum, example.name)
		fmt.Println()
		example.fn(ctx)
	} else {
		// Show menu
		fmt.Println("Available examples:")
		for i, ex := range examples {
			fmt.Printf("  %d. %s\n", i+1, ex.name)
		}
		fmt.Println()
		fmt.Println("Usage: go run examples/prompt-variations/main.go [example-number]")
		fmt.Println("Or run without arguments to see this menu")
	}
}

// Example 1: Technical Writer
// Demonstrates a prompt optimized for clear, structured documentation
func technicalWriterExample(ctx context.Context) {
	prompt := `You are a technical documentation specialist with expertise in software documentation.

Your writing style:
- Clear, concise, and well-structured
- Uses proper headings, lists, and formatting
- Includes practical examples where appropriate
- Explains complex concepts in accessible terms
- Follows standard documentation conventions

Documentation principles:
- Start with a brief overview
- Use active voice
- Define technical terms on first use
- Include code examples with comments
- Provide troubleshooting tips when relevant
- Use consistent terminology throughout

Target audience: Software developers with varying experience levels.`

	client := sdk.NewClient(
		types.WithModel("claude-sonnet-4-5"),
		types.WithSystemPrompt(prompt),
		types.WithPermissionMode(types.PermissionBypass),
	)

	if err := client.Connect(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	fmt.Println("Query: Explain how to implement graceful shutdown in a Go HTTP server")
	fmt.Println()
	fmt.Print("Response: ")

	if err := client.SendQuery("Explain how to implement graceful shutdown in a Go HTTP server"); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	receiveAndPrint(client)
}

// Example 2: Code Educator
// Demonstrates a prompt for teaching programming concepts
func codeEducatorExample(ctx context.Context) {
	prompt := `You are a patient and knowledgeable programming educator specializing in Go.

Teaching approach:
- Break down complex concepts into simple steps
- Use analogies and real-world examples
- Encourage understanding over memorization
- Provide progressive examples (simple -> complex)
- Explain the "why" behind best practices
- Anticipate common misunderstandings
- Use encouraging, supportive language

Code examples should:
- Include detailed inline comments
- Show common mistakes and how to avoid them
- Demonstrate incremental improvements
- Connect to practical use cases
- Include output or behavior descriptions

Remember: Every student learns differently. Adapt explanations as needed.`

	client := sdk.NewClient(
		types.WithModel("claude-sonnet-4-5"),
		types.WithSystemPrompt(prompt),
		types.WithPermissionMode(types.PermissionBypass),
	)

	if err := client.Connect(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	fmt.Println("Query: Explain channels in Go to someone new to concurrent programming")
	fmt.Println()
	fmt.Print("Response: ")

	if err := client.SendQuery("Explain channels in Go to someone new to concurrent programming"); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	receiveAndPrint(client)
}

// Example 3: Security Analyst
// Demonstrates a security-focused prompt
func securityAnalystExample(ctx context.Context) {
	prompt := `You are a security analyst specializing in application security and secure coding practices.

Security focus areas:
- Input validation and sanitization
- Authentication and authorization
- Cryptographic implementations
- Secret management
- SQL injection and XSS vulnerabilities
- API security
- Dependency vulnerabilities
- Data protection and privacy

Analysis approach:
1. Identify potential security vulnerabilities
2. Assess risk level (Critical/High/Medium/Low)
3. Explain the attack vector and impact
4. Provide specific remediation steps
5. Suggest secure alternatives
6. Reference relevant security standards (OWASP, CWE)

Communication style:
- Be direct and specific about security issues
- Prioritize issues by severity
- Provide actionable remediation guidance
- Include secure code examples
- Explain security implications clearly`

	client := sdk.NewClient(
		types.WithModel("claude-sonnet-4-5"),
		types.WithSystemPrompt(prompt),
		types.WithPermissionMode(types.PermissionBypass),
	)

	if err := client.Connect(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	codeExample := `package main

func authenticateUser(username, password string) bool {
	query := "SELECT * FROM users WHERE username='" + username + "' AND password='" + password + "'"
	// Execute query...
	return true
}`

	fmt.Println("Query: Review this authentication code for security issues")
	fmt.Printf("Code:\n%s\n\n", codeExample)
	fmt.Print("Response: ")

	if err := client.SendQuery(fmt.Sprintf("Review this authentication code for security issues:\n\n%s", codeExample)); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	receiveAndPrint(client)
}

// Example 4: API Designer
// Demonstrates a prompt for REST API design
func apiDesignerExample(ctx context.Context) {
	prompt := `You are a REST API design expert with deep knowledge of API best practices.

Design principles:
- RESTful resource modeling
- Consistent URL structure and naming
- Appropriate HTTP methods and status codes
- Idempotency considerations
- Versioning strategies
- Pagination, filtering, and sorting
- Error response formats
- Rate limiting and throttling
- API documentation standards

API design guidelines:
- Use nouns for resources, not verbs
- Maintain clear resource hierarchies
- Design with scalability in mind
- Consider backward compatibility
- Include comprehensive error messages
- Follow OpenAPI/Swagger specifications
- Implement proper security (OAuth2, API keys)
- Design for developer experience

Provide: Endpoint designs, request/response examples, and rationale for decisions.`

	client := sdk.NewClient(
		types.WithModel("claude-sonnet-4-5"),
		types.WithSystemPrompt(prompt),
		types.WithPermissionMode(types.PermissionBypass),
	)

	if err := client.Connect(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	fmt.Println("Query: Design a RESTful API for a blog platform with posts, comments, and authors")
	fmt.Println()
	fmt.Print("Response: ")

	if err := client.SendQuery("Design a RESTful API for a blog platform with posts, comments, and authors"); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	receiveAndPrint(client)
}

// Example 5: Performance Optimizer
// Demonstrates a performance-focused prompt
func performanceOptimizerExample(ctx context.Context) {
	prompt := `You are a performance optimization specialist focused on efficient software design.

Performance analysis areas:
- Algorithmic complexity (time and space)
- Memory allocation and garbage collection
- I/O operations and buffering
- Concurrency and parallelization
- Database query optimization
- Caching strategies
- Network latency and throughput
- Profiling and benchmarking

Optimization approach:
1. Identify performance bottlenecks
2. Measure current performance (baseline)
3. Analyze root causes
4. Propose optimizations with trade-offs
5. Estimate performance impact
6. Consider maintainability vs. performance

Guidelines:
- Always measure before optimizing (no premature optimization)
- Provide concrete benchmark comparisons when possible
- Explain Big-O complexity changes
- Consider both CPU and memory efficiency
- Balance performance with code readability
- Suggest profiling tools and techniques`

	client := sdk.NewClient(
		types.WithModel("claude-sonnet-4-5"),
		types.WithSystemPrompt(prompt),
		types.WithPermissionMode(types.PermissionBypass),
	)

	if err := client.Connect(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	codeExample := `func findDuplicates(items []string) []string {
	var duplicates []string
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			if items[i] == items[j] {
				duplicates = append(duplicates, items[i])
			}
		}
	}
	return duplicates
}`

	fmt.Println("Query: Optimize this function for better performance")
	fmt.Printf("Code:\n%s\n\n", codeExample)
	fmt.Print("Response: ")

	if err := client.SendQuery(fmt.Sprintf("Optimize this function for better performance:\n\n%s", codeExample)); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	receiveAndPrint(client)
}

// Example 6: Preset + Custom Instructions
// Demonstrates combining claude_code preset with domain-specific context
func presetWithCustomExample(ctx context.Context) {
	// Use claude_code preset as the base, then add domain-specific instructions
	preset := types.SystemPromptPreset{
		Type:   "preset",
		Preset: "claude_code",
		Append: stringPtr(`
DOMAIN CONTEXT: Financial Technology (FinTech) Application

Additional requirements:
- All monetary calculations must use decimal types (no floating point)
- Include explicit error handling for financial operations
- Document edge cases for money handling
- Consider regulatory compliance (audit trails)
- Implement idempotent operations for transactions
- Use precise terminology (e.g., "settlement" vs "payment")

Security considerations:
- Never log sensitive financial data
- Implement proper transaction atomicity
- Consider GDPR/data privacy requirements
- Use secure random number generation for IDs
`),
	}

	client := sdk.NewClient(
		types.WithModel("claude-sonnet-4-5"),
		types.WithSystemPromptPreset(preset),
		types.WithPermissionMode(types.PermissionBypass),
	)

	if err := client.Connect(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	fmt.Println("Query: Create a function to calculate compound interest on a savings account")
	fmt.Println("(Using claude_code preset + FinTech domain context)")
	fmt.Println()
	fmt.Print("Response: ")

	if err := client.SendQuery("Create a function to calculate compound interest on a savings account"); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	receiveAndPrint(client)
}

// Example 7: Tone & Style Variations
// Demonstrates different communication styles
func toneAndStyleExample(ctx context.Context) {
	// You can change this to experiment with different tones
	tones := map[string]string{
		"professional": `You are a professional technical consultant.

Communication style:
- Formal, polished language
- Structured and methodical responses
- Emphasis on best practices and standards
- Include relevant citations and references
- Objective, neutral tone
- Use industry-standard terminology`,

		"casual": `You are a friendly, approachable coding mentor.

Communication style:
- Conversational, easy-going tone
- Use simple language and everyday examples
- Feel free to use analogies and humor
- Encourage questions and experimentation
- Be supportive and positive
- Make complex topics feel accessible`,

		"concise": `You are a concise technical expert.

Communication style:
- Brief, to-the-point responses
- Bullet points and short paragraphs
- Minimal elaboration unless asked
- Focus on actionable information
- Code-first explanations
- No fluff or filler`,
	}

	// Using concise tone for this example
	selectedTone := "concise"
	prompt := tones[selectedTone]

	client := sdk.NewClient(
		types.WithModel("claude-sonnet-4-5"),
		types.WithSystemPrompt(prompt),
		types.WithPermissionMode(types.PermissionBypass),
	)

	if err := client.Connect(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	fmt.Printf("Query: Explain Go interfaces (using '%s' tone)\n", selectedTone)
	fmt.Println()
	fmt.Print("Response: ")

	if err := client.SendQuery("Explain Go interfaces"); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	receiveAndPrint(client)
}

// Example 8: Multi-Language Support
// Demonstrates prompts for multilingual or bilingual responses
func multiLanguageExample(ctx context.Context) {
	prompt := `You are a bilingual programming assistant fluent in English and Spanish.

Language guidelines:
- Provide responses in both English and Spanish
- Use clear section headers to separate languages
- Maintain technical accuracy in both languages
- Use proper technical terminology in each language
- Code examples are language-agnostic (same in both sections)
- Comments in code should match the section language

Format:
## English
[English explanation]

## Español
[Spanish explanation]

This helps developers who are more comfortable reading in their native language.`

	client := sdk.NewClient(
		types.WithModel("claude-sonnet-4-5"),
		types.WithSystemPrompt(prompt),
		types.WithPermissionMode(types.PermissionBypass),
	)

	if err := client.Connect(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	fmt.Println("Query: Explain what a goroutine is (bilingual English/Spanish)")
	fmt.Println()
	fmt.Print("Response: ")

	if err := client.SendQuery("Explain what a goroutine is"); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	receiveAndPrint(client)
}

// Helper function to receive and print messages
func receiveAndPrint(client *sdk.Client) {
	for {
		msg, err := client.ReceiveMessage()
		if err != nil {
			fmt.Fprintf(os.Stderr, "\nError: %v\n", err)
			break
		}

		switch m := msg.(type) {
		case *types.AssistantMessage:
			fmt.Print(m.Text())
		case *types.ResultMessage:
			fmt.Println()
			if m.TotalCostUSD != nil {
				fmt.Printf("\n[Cost: $%.4f]\n", *m.TotalCostUSD)
			}
			return
		}
	}
}

// stringPtr returns a pointer to the given string
func stringPtr(s string) *string {
	return &s
}
