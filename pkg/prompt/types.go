package prompt

import "errors"

// Common errors
var (
	// ErrInterrupted is returned when user interrupts the prompt (e.g., Ctrl+C)
	ErrInterrupted = errors.New("prompt interrupted")

	// ErrNoOptions is returned when a selection prompt has no options
	ErrNoOptions = errors.New("no options provided")

	// ErrInvalidInput is returned when user provides invalid input
	ErrInvalidInput = errors.New("invalid input")

	// ErrValidationFailed is returned when input validation fails
	ErrValidationFailed = errors.New("validation failed")
)

// PromptConfig holds configuration for prompts
type PromptConfig struct {
	// NoColor disables colored output
	NoColor bool

	// Quiet mode suppresses prompts and uses defaults
	Quiet bool

	// NonInteractive mode returns errors instead of prompting
	NonInteractive bool
}

// Option is a functional option for configuring prompts
type Option func(*PromptConfig)

// WithNoColor disables colored output
func WithNoColor() Option {
	return func(c *PromptConfig) {
		c.NoColor = true
	}
}

// WithQuiet enables quiet mode (use defaults)
func WithQuiet() Option {
	return func(c *PromptConfig) {
		c.Quiet = true
	}
}

// WithNonInteractive enables non-interactive mode
func WithNonInteractive() Option {
	return func(c *PromptConfig) {
		c.NonInteractive = true
	}
}

// SelectOption represents an option in a selection prompt
type SelectOption struct {
	// Label is the display text for the option
	Label string

	// Value is the underlying value
	Value interface{}

	// Description provides additional context
	Description string

	// Disabled prevents selection of this option
	Disabled bool
}

// InputConfig holds configuration for input prompts
type InputConfig struct {
	// Required indicates if empty input is allowed
	Required bool

	// Hidden masks input (for passwords)
	Hidden bool

	// MultiLine allows multi-line input
	MultiLine bool

	// Placeholder shows hint text
	Placeholder string

	// MaxLength limits input length
	MaxLength int

	// MinLength requires minimum length
	MinLength int
}

// ConfirmConfig holds configuration for confirmation prompts
type ConfirmConfig struct {
	// Destructive shows warning for dangerous operations
	Destructive bool

	// RequireExplicit requires typing "yes" instead of y/n
	RequireExplicit bool

	// Warning message to display before prompt
	Warning string
}
