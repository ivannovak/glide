package prompt

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPromptConfig(t *testing.T) {
	config := &PromptConfig{}

	// Test functional options
	WithNoColor()(config)
	assert.True(t, config.NoColor)

	WithQuiet()(config)
	assert.True(t, config.Quiet)

	WithNonInteractive()(config)
	assert.True(t, config.NonInteractive)
}

func TestSelectOption(t *testing.T) {
	option := SelectOption{
		Label:       "Test Option",
		Value:       42,
		Description: "A test option",
		Disabled:    false,
	}

	assert.Equal(t, "Test Option", option.Label)
	assert.Equal(t, 42, option.Value)
	assert.Equal(t, "A test option", option.Description)
	assert.False(t, option.Disabled)
}

func TestInputConfig(t *testing.T) {
	config := InputConfig{
		Required:    true,
		Hidden:      false,
		MultiLine:   false,
		Placeholder: "Enter value",
		MaxLength:   100,
		MinLength:   5,
	}

	assert.True(t, config.Required)
	assert.False(t, config.Hidden)
	assert.False(t, config.MultiLine)
	assert.Equal(t, "Enter value", config.Placeholder)
	assert.Equal(t, 100, config.MaxLength)
	assert.Equal(t, 5, config.MinLength)
}

func TestConfirmConfig(t *testing.T) {
	config := ConfirmConfig{
		Destructive:     true,
		RequireExplicit: true,
		Warning:         "This is dangerous",
	}

	assert.True(t, config.Destructive)
	assert.True(t, config.RequireExplicit)
	assert.Equal(t, "This is dangerous", config.Warning)
}

func TestErrorConstants(t *testing.T) {
	assert.Equal(t, "prompt interrupted", ErrInterrupted.Error())
	assert.Equal(t, "no options provided", ErrNoOptions.Error())
	assert.Equal(t, "invalid input", ErrInvalidInput.Error())
	assert.Equal(t, "validation failed", ErrValidationFailed.Error())
}

func TestNewPrompter(t *testing.T) {
	prompter := New()

	assert.NotNil(t, prompter)
	assert.NotNil(t, prompter.reader)
	assert.NotNil(t, prompter.writer)
}

// Mock prompter for testing
type MockPrompter struct {
	ConfirmResponses  []bool
	ConfirmErrors     []error
	SelectResponses   []int
	SelectValues      []string
	SelectErrors      []error
	InputResponses    []string
	InputErrors       []error
	PasswordResponses []string
	PasswordErrors    []error

	confirmIndex  int
	selectIndex   int
	inputIndex    int
	passwordIndex int
}

func NewMockPrompter() *MockPrompter {
	return &MockPrompter{}
}

func (m *MockPrompter) Confirm(message string, defaultValue bool) (bool, error) {
	if m.confirmIndex >= len(m.ConfirmResponses) {
		return defaultValue, fmt.Errorf("no more confirm responses")
	}

	response := m.ConfirmResponses[m.confirmIndex]
	var err error
	if m.confirmIndex < len(m.ConfirmErrors) {
		err = m.ConfirmErrors[m.confirmIndex]
	}

	m.confirmIndex++
	return response, err
}

func (m *MockPrompter) Select(message string, options []string, defaultIndex int) (int, string, error) {
	if m.selectIndex >= len(m.SelectResponses) {
		return defaultIndex, options[defaultIndex], fmt.Errorf("no more select responses")
	}

	index := m.SelectResponses[m.selectIndex]
	value := ""
	if m.selectIndex < len(m.SelectValues) {
		value = m.SelectValues[m.selectIndex]
	} else if index >= 0 && index < len(options) {
		value = options[index]
	}

	var err error
	if m.selectIndex < len(m.SelectErrors) {
		err = m.SelectErrors[m.selectIndex]
	}

	m.selectIndex++
	return index, value, err
}

func (m *MockPrompter) Input(message string, defaultValue string, validator InputValidator) (string, error) {
	if m.inputIndex >= len(m.InputResponses) {
		return defaultValue, fmt.Errorf("no more input responses")
	}

	response := m.InputResponses[m.inputIndex]
	var err error
	if m.InputErrors != nil && m.inputIndex < len(m.InputErrors) {
		err = m.InputErrors[m.inputIndex]
	}

	// Apply validator if provided
	if validator != nil && err == nil {
		if validationErr := validator(response); validationErr != nil {
			err = validationErr
		}
	}

	m.inputIndex++
	return response, err
}

func (m *MockPrompter) Password(message string) (string, error) {
	if m.passwordIndex >= len(m.PasswordResponses) {
		return "", fmt.Errorf("no more password responses")
	}

	response := m.PasswordResponses[m.passwordIndex]
	var err error
	if m.passwordIndex < len(m.PasswordErrors) {
		err = m.PasswordErrors[m.passwordIndex]
	}

	m.passwordIndex++
	return response, err
}

func TestMockPrompterConfirm(t *testing.T) {
	mock := NewMockPrompter()
	mock.ConfirmResponses = []bool{true, false}
	mock.ConfirmErrors = []error{nil, fmt.Errorf("confirm error")}

	// First call - success
	result, err := mock.Confirm("Test?", false)
	assert.True(t, result)
	assert.NoError(t, err)

	// Second call - error
	result, err = mock.Confirm("Test?", true)
	assert.False(t, result)
	assert.Error(t, err)
	assert.Equal(t, "confirm error", err.Error())
}

func TestMockPrompterSelect(t *testing.T) {
	mock := NewMockPrompter()
	mock.SelectResponses = []int{1, 0}
	mock.SelectValues = []string{"option2", "option1"}
	mock.SelectErrors = []error{nil, fmt.Errorf("select error")}

	options := []string{"option1", "option2", "option3"}

	// First call - success
	index, value, err := mock.Select("Choose:", options, 0)
	assert.Equal(t, 1, index)
	assert.Equal(t, "option2", value)
	assert.NoError(t, err)

	// Second call - error
	index, value, err = mock.Select("Choose:", options, 2)
	assert.Equal(t, 0, index)
	assert.Equal(t, "option1", value)
	assert.Error(t, err)
	assert.Equal(t, "select error", err.Error())
}

func TestMockPrompterInput(t *testing.T) {
	mock := NewMockPrompter()
	mock.InputResponses = []string{"test input", "invalid"}
	mock.InputErrors = []error{nil, fmt.Errorf("input error")}

	// First call - success
	result, err := mock.Input("Enter:", "default", nil)
	assert.Equal(t, "test input", result)
	assert.NoError(t, err)

	// Second call - error
	result, err = mock.Input("Enter:", "default", nil)
	assert.Equal(t, "invalid", result)
	assert.Error(t, err)
	assert.Equal(t, "input error", err.Error())
}

func TestMockPrompterInputWithValidator(t *testing.T) {
	validator := func(input string) error {
		if len(input) < 5 {
			return fmt.Errorf("too short")
		}
		return nil
	}

	// Test validation passes (simpler test)
	mock := NewMockPrompter()
	mock.InputResponses = []string{"long enough"}

	result, err := mock.Input("Enter:", "default", validator)
	assert.Equal(t, "long enough", result)
	assert.NoError(t, err)
}

func TestMockPrompterPassword(t *testing.T) {
	mock := NewMockPrompter()
	mock.PasswordResponses = []string{"secret123", ""}
	mock.PasswordErrors = []error{nil, fmt.Errorf("password error")}

	// First call - success
	result, err := mock.Password("Enter password:")
	assert.Equal(t, "secret123", result)
	assert.NoError(t, err)

	// Second call - error
	result, err = mock.Password("Enter password:")
	assert.Equal(t, "", result)
	assert.Error(t, err)
	assert.Equal(t, "password error", err.Error())
}

func TestInputValidator(t *testing.T) {
	// Test validator function type
	validator := func(input string) error {
		if strings.TrimSpace(input) == "" {
			return ErrInvalidInput
		}
		return nil
	}

	assert.NoError(t, validator("valid input"))
	assert.Equal(t, ErrInvalidInput, validator(""))
	assert.Equal(t, ErrInvalidInput, validator("   "))
}

func TestPrompterInterface(t *testing.T) {
	// Test that our mock implements the interface
	var prompter Prompter = NewMockPrompter()

	// This should compile if interface is satisfied
	assert.NotNil(t, prompter)
}

// Test validators

func TestRequiredValidator(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:        "valid input",
			input:       "test value",
			expectError: false,
		},
		{
			name:        "empty string",
			input:       "",
			expectError: true,
		},
		{
			name:        "whitespace only",
			input:       "   ",
			expectError: true,
		},
		{
			name:        "tab only",
			input:       "\t",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RequiredValidator(tt.input)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "required")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMinLengthValidator(t *testing.T) {
	validator := MinLengthValidator(5)

	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:        "meets minimum",
			input:       "12345",
			expectError: false,
		},
		{
			name:        "exceeds minimum",
			input:       "123456",
			expectError: false,
		},
		{
			name:        "below minimum",
			input:       "1234",
			expectError: true,
		},
		{
			name:        "empty",
			input:       "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator(tt.input)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "at least 5")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMaxLengthValidator(t *testing.T) {
	validator := MaxLengthValidator(10)

	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:        "within limit",
			input:       "12345",
			expectError: false,
		},
		{
			name:        "at limit",
			input:       "1234567890",
			expectError: false,
		},
		{
			name:        "exceeds limit",
			input:       "12345678901",
			expectError: true,
		},
		{
			name:        "empty",
			input:       "",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator(tt.input)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "at most 10")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPathValidator(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		errorMsg    string
		expectError bool
	}{
		{
			name:        "valid unix path",
			input:       "/usr/local/bin",
			expectError: false,
		},
		{
			name:        "valid relative path",
			input:       "./test/path",
			expectError: false,
		},
		{
			name:        "valid windows path",
			input:       "C:\\Users\\test",
			expectError: false,
		},
		{
			name:        "empty path",
			input:       "",
			expectError: true,
			errorMsg:    "cannot be empty",
		},
		{
			name:        "null byte in path",
			input:       "/path/with\x00null",
			expectError: true,
			errorMsg:    "invalid characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := PathValidator(tt.input)
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestChainValidators(t *testing.T) {
	validator := ChainValidators(
		RequiredValidator,
		MinLengthValidator(3),
		MaxLengthValidator(10),
	)

	tests := []struct {
		name          string
		input         string
		errorContains string
		expectError   bool
	}{
		{
			name:        "valid input",
			input:       "test",
			expectError: false,
		},
		{
			name:          "empty fails required",
			input:         "",
			expectError:   true,
			errorContains: "required",
		},
		{
			name:          "too short fails min length",
			input:         "ab",
			expectError:   true,
			errorContains: "at least 3",
		},
		{
			name:          "too long fails max length",
			input:         "12345678901",
			expectError:   true,
			errorContains: "at most 10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator(tt.input)
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestChainValidators_Empty(t *testing.T) {
	// Empty chain should succeed for any input
	validator := ChainValidators()

	err := validator("anything")
	assert.NoError(t, err)
}

func TestChainValidators_Order(t *testing.T) {
	// Validators should run in order and stop at first failure
	firstErr := fmt.Errorf("first validator failed")
	secondErr := fmt.Errorf("second validator failed")

	firstValidator := func(input string) error {
		if input == "fail-first" {
			return firstErr
		}
		return nil
	}

	secondValidator := func(input string) error {
		if input == "fail-second" {
			return secondErr
		}
		return nil
	}

	validator := ChainValidators(firstValidator, secondValidator)

	// Test first validator fails
	err := validator("fail-first")
	assert.Equal(t, firstErr, err)

	// Test second validator fails
	err = validator("fail-second")
	assert.Equal(t, secondErr, err)

	// Test both pass
	err = validator("pass")
	assert.NoError(t, err)
}

func TestMinLengthValidator_Zero(t *testing.T) {
	validator := MinLengthValidator(0)

	err := validator("")
	assert.NoError(t, err)
}

func TestMaxLengthValidator_Zero(t *testing.T) {
	validator := MaxLengthValidator(0)

	err := validator("")
	assert.NoError(t, err)

	err = validator("a")
	assert.Error(t, err)
}

func TestPathValidator_EdgeCases(t *testing.T) {
	// Test various edge cases
	tests := []struct {
		name  string
		path  string
		valid bool
	}{
		{"single dot", ".", true},
		{"double dot", "..", true},
		{"tilde", "~/test", true},
		{"dollar sign", "$HOME/test", true},
		{"space", "/path with space", true},
		{"special chars", "/path-with_special.chars", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := PathValidator(tt.path)
			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
