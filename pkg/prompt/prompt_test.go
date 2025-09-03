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
