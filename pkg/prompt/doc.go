// Package prompt provides interactive terminal prompts for user input.
//
// This package offers various prompt types for gathering user input in
// an interactive terminal environment. It handles TTY detection, color
// output, and graceful degradation for non-interactive environments.
//
// # Text Input
//
// Get text input from the user:
//
//	prompter := prompt.New()
//
//	// Simple text prompt
//	name, err := prompter.Input("Enter your name:", "")
//
//	// With default value
//	host, err := prompter.Input("Database host:", "localhost")
//
// # Confirmation Prompts
//
// Ask yes/no questions:
//
//	proceed, err := prompter.Confirm("Continue with installation?", true)
//	if proceed {
//	    // User confirmed
//	}
//
// # Selection Prompts
//
// Let users choose from options:
//
//	options := []string{"Development", "Staging", "Production"}
//	choice, err := prompter.Select("Choose environment:", options)
//	// choice is the selected string
//
// # Multi-Select
//
// Allow multiple selections:
//
//	options := []string{"API", "Database", "Cache", "Queue"}
//	selected, err := prompter.MultiSelect("Select components:", options)
//	// selected is []string of chosen options
//
// # Password Input
//
// Secure password entry (masked):
//
//	password, err := prompter.Password("Enter password:")
//
// # Non-Interactive Mode
//
// Handle non-TTY environments gracefully:
//
//	prompter := prompt.New(prompt.WithNonInteractive())
//	// Uses default values without prompting
//
// # Color Support
//
// Prompts automatically detect and use terminal colors:
//
//	prompter := prompt.New(prompt.WithNoColor())
//	// Disables color output
//
// # Integration with Container
//
// The container can provide a configured prompter:
//
//	c.Run(ctx, func(p *prompt.Prompter) error {
//	    name, err := p.Input("Name:", "")
//	    return nil
//	})
package prompt
