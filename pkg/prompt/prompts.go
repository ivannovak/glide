package prompt

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

// Prompter interface for testing
type Prompter interface {
	Confirm(message string, defaultValue bool) (bool, error)
	Select(message string, options []string, defaultIndex int) (int, string, error)
	Input(message string, defaultValue string, validator InputValidator) (string, error)
	Password(message string) (string, error)
}

// InputValidator is a function type for validating user input
type InputValidator func(input string) error

// DefaultPrompter implements the Prompter interface using stdin/stdout
type DefaultPrompter struct {
	reader *bufio.Reader
	writer *os.File
}

// New creates a new DefaultPrompter
func New() *DefaultPrompter {
	return &DefaultPrompter{
		reader: bufio.NewReader(os.Stdin),
		writer: os.Stdout,
	}
}

// Confirm displays a yes/no confirmation prompt
func (p *DefaultPrompter) Confirm(message string, defaultValue bool) (bool, error) {
	defaultStr := "y/N"
	if defaultValue {
		defaultStr = "Y/n"
	}

	// Format the prompt
	prompt := fmt.Sprintf("%s %s [%s]: ", 
		color.YellowString("?"),
		message,
		defaultStr,
	)

	fmt.Fprint(p.writer, prompt)

	// Read user input
	input, err := p.reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("failed to read input: %w", err)
	}

	// Trim whitespace
	input = strings.TrimSpace(strings.ToLower(input))

	// Handle empty input (use default)
	if input == "" {
		return defaultValue, nil
	}

	// Parse response
	switch input {
	case "y", "yes", "true", "1":
		return true, nil
	case "n", "no", "false", "0":
		return false, nil
	default:
		// If invalid input, return default
		return defaultValue, nil
	}
}

// Select displays a selection prompt with options
func (p *DefaultPrompter) Select(message string, options []string, defaultIndex int) (int, string, error) {
	if len(options) == 0 {
		return -1, "", fmt.Errorf("no options provided")
	}

	// Validate default index
	if defaultIndex < 0 || defaultIndex >= len(options) {
		defaultIndex = 0
	}

	// Display the prompt message
	fmt.Fprintf(p.writer, "%s %s\n", 
		color.YellowString("?"),
		message,
	)

	// Display options
	for i, option := range options {
		prefix := "  "
		if i == defaultIndex {
			prefix = color.CyanString("❯ ")
		}
		fmt.Fprintf(p.writer, "%s%d) %s\n", prefix, i+1, option)
	}

	// Show input prompt
	fmt.Fprintf(p.writer, "\n%s Enter choice [1-%d] (default: %d): ", 
		color.YellowString("›"),
		len(options),
		defaultIndex+1,
	)

	// Read user input
	input, err := p.reader.ReadString('\n')
	if err != nil {
		return -1, "", fmt.Errorf("failed to read input: %w", err)
	}

	input = strings.TrimSpace(input)

	// Handle empty input (use default)
	if input == "" {
		return defaultIndex, options[defaultIndex], nil
	}

	// Parse selection
	var choice int
	_, err = fmt.Sscanf(input, "%d", &choice)
	if err != nil {
		// Try to match by string
		inputLower := strings.ToLower(input)
		for i, option := range options {
			if strings.ToLower(option) == inputLower || 
			   strings.HasPrefix(strings.ToLower(option), inputLower) {
				return i, option, nil
			}
		}
		// Invalid input, use default
		return defaultIndex, options[defaultIndex], nil
	}

	// Validate choice (1-based from user perspective)
	choice--
	if choice < 0 || choice >= len(options) {
		return defaultIndex, options[defaultIndex], nil
	}

	return choice, options[choice], nil
}

// Input displays a text input prompt with optional validation
func (p *DefaultPrompter) Input(message string, defaultValue string, validator InputValidator) (string, error) {
	// Format the prompt
	defaultStr := ""
	if defaultValue != "" {
		defaultStr = fmt.Sprintf(" (default: %s)", defaultValue)
	}

	prompt := fmt.Sprintf("%s %s%s: ", 
		color.YellowString("?"),
		message,
		defaultStr,
	)

	for {
		fmt.Fprint(p.writer, prompt)

		// Read user input
		input, err := p.reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("failed to read input: %w", err)
		}

		// Trim whitespace
		input = strings.TrimSpace(input)

		// Handle empty input (use default)
		if input == "" && defaultValue != "" {
			input = defaultValue
		}

		// Validate input if validator provided
		if validator != nil {
			if err := validator(input); err != nil {
				fmt.Fprintf(p.writer, "%s %s\n", 
					color.RedString("✗"),
					err.Error(),
				)
				continue // Ask again
			}
		}

		return input, nil
	}
}

// Password displays a password input prompt (note: input is visible in terminal)
func (p *DefaultPrompter) Password(message string) (string, error) {
	// Note: For production use, consider using golang.org/x/term for hidden input
	fmt.Fprintf(p.writer, "%s %s: ", 
		color.YellowString("?"),
		message,
	)

	input, err := p.reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read input: %w", err)
	}

	return strings.TrimSpace(input), nil
}

// Common validators

// RequiredValidator ensures the input is not empty
func RequiredValidator(input string) error {
	if strings.TrimSpace(input) == "" {
		return fmt.Errorf("value is required")
	}
	return nil
}

// MinLengthValidator ensures the input meets minimum length
func MinLengthValidator(min int) InputValidator {
	return func(input string) error {
		if len(input) < min {
			return fmt.Errorf("must be at least %d characters", min)
		}
		return nil
	}
}

// MaxLengthValidator ensures the input doesn't exceed maximum length
func MaxLengthValidator(max int) InputValidator {
	return func(input string) error {
		if len(input) > max {
			return fmt.Errorf("must be at most %d characters", max)
		}
		return nil
	}
}

// PathValidator ensures the input is a valid path
func PathValidator(input string) error {
	if input == "" {
		return fmt.Errorf("path cannot be empty")
	}
	
	// Check for invalid characters
	if strings.ContainsAny(input, "\x00") {
		return fmt.Errorf("path contains invalid characters")
	}
	
	return nil
}

// ChainValidators combines multiple validators
func ChainValidators(validators ...InputValidator) InputValidator {
	return func(input string) error {
		for _, validator := range validators {
			if err := validator(input); err != nil {
				return err
			}
		}
		return nil
	}
}

// Convenience functions using the default prompter

var defaultPrompter = New()

// Confirm is a convenience function using the default prompter
func Confirm(message string, defaultValue bool) (bool, error) {
	return defaultPrompter.Confirm(message, defaultValue)
}

// Select is a convenience function using the default prompter
func Select(message string, options []string, defaultIndex int) (int, string, error) {
	return defaultPrompter.Select(message, options, defaultIndex)
}

// Input is a convenience function using the default prompter
func Input(message string, defaultValue string, validator InputValidator) (string, error) {
	return defaultPrompter.Input(message, defaultValue, validator)
}

// Password is a convenience function using the default prompter
func Password(message string) (string, error) {
	return defaultPrompter.Password(message)
}

// ConfirmDestructive displays a confirmation prompt for destructive operations
// It requires explicit confirmation and shows a warning
func ConfirmDestructive(operation string) (bool, error) {
	fmt.Fprintf(os.Stdout, "\n%s This is a destructive operation!\n", 
		color.RedString("⚠"),
	)
	
	message := fmt.Sprintf("Are you sure you want to %s?", operation)
	
	// Default to false for destructive operations
	confirmed, err := Confirm(message, false)
	if err != nil {
		return false, err
	}
	
	if !confirmed {
		fmt.Fprintf(os.Stdout, "%s Operation cancelled\n", 
			color.YellowString("→"),
		)
	}
	
	return confirmed, nil
}

// SelectProject displays a project selection prompt
func SelectProject(projects []string, current string) (string, error) {
	if len(projects) == 0 {
		return "", fmt.Errorf("no projects available")
	}
	
	// Find current project index
	currentIndex := 0
	for i, p := range projects {
		if p == current {
			currentIndex = i
			break
		}
	}
	
	_, selected, err := Select("Select a project", projects, currentIndex)
	return selected, err
}

// InputPath prompts for a file or directory path with validation
func InputPath(message string, defaultPath string) (string, error) {
	return Input(message, defaultPath, PathValidator)
}