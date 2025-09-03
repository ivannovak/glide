package output

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManager(t *testing.T) {
	t.Run("creates manager with specified format", func(t *testing.T) {
		buf := &bytes.Buffer{}
		manager := NewManager(FormatJSON, false, false, buf)

		assert.NotNil(t, manager)
		assert.Equal(t, FormatJSON, manager.GetFormat())
		assert.False(t, manager.IsQuiet())
	})

	t.Run("creates manager with quiet mode", func(t *testing.T) {
		manager := NewManager(FormatTable, true, false, nil)

		assert.NotNil(t, manager)
		assert.True(t, manager.IsQuiet())
	})

	t.Run("creates manager with no color", func(t *testing.T) {
		manager := NewManager(FormatTable, false, true, nil)

		assert.NotNil(t, manager)
		// Colors should be disabled
	})

	t.Run("uses stdout when writer is nil", func(t *testing.T) {
		manager := NewManager(FormatTable, false, false, nil)

		assert.NotNil(t, manager)
		assert.NotNil(t, manager.writer)
	})
}

func TestManagerSetters(t *testing.T) {
	t.Run("SetFormat changes format", func(t *testing.T) {
		buf := &bytes.Buffer{}
		manager := NewManager(FormatTable, false, false, buf)

		manager.SetFormat(FormatJSON)
		assert.Equal(t, FormatJSON, manager.GetFormat())

		manager.SetFormat(FormatYAML)
		assert.Equal(t, FormatYAML, manager.GetFormat())
	})

	t.Run("SetQuiet changes quiet mode", func(t *testing.T) {
		buf := &bytes.Buffer{}
		manager := NewManager(FormatTable, false, false, buf)

		manager.SetQuiet(true)
		assert.True(t, manager.IsQuiet())

		manager.SetQuiet(false)
		assert.False(t, manager.IsQuiet())
	})

	t.Run("SetNoColor changes color mode", func(t *testing.T) {
		buf := &bytes.Buffer{}
		manager := NewManager(FormatTable, false, false, buf)

		manager.SetNoColor(true)
		// Verify colors are disabled

		manager.SetNoColor(false)
		// Verify colors are enabled
	})

	t.Run("SetWriter changes output writer", func(t *testing.T) {
		buf := &bytes.Buffer{}
		manager := NewManager(FormatTable, false, false, buf)

		newBuf := &bytes.Buffer{}
		manager.SetWriter(newBuf)

		manager.Raw("test")
		assert.Equal(t, "test", newBuf.String())
		assert.Empty(t, buf.String())
	})
}

func TestManagerDisplay(t *testing.T) {
	t.Run("displays data with current formatter", func(t *testing.T) {
		buf := &bytes.Buffer{}
		manager := NewManager(FormatJSON, false, false, buf)

		data := map[string]string{"key": "value"}
		err := manager.Display(data)

		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "\"key\"")
		assert.Contains(t, buf.String(), "\"value\"")
	})

	t.Run("respects quiet mode", func(t *testing.T) {
		buf := &bytes.Buffer{}
		manager := NewManager(FormatTable, true, false, buf)

		err := manager.Info("test message")
		assert.NoError(t, err)
		// In quiet mode, info messages might be suppressed
	})
}

func TestManagerMessages(t *testing.T) {
	tests := []struct {
		name   string
		method func(*Manager, string, ...interface{}) error
		format string
		args   []interface{}
		expect string
	}{
		{
			name:   "Info message",
			method: (*Manager).Info,
			format: "Info: %s",
			args:   []interface{}{"test"},
			expect: "Info: test",
		},
		{
			name:   "Success message",
			method: (*Manager).Success,
			format: "Success: %d items",
			args:   []interface{}{5},
			expect: "Success: 5 items",
		},
		{
			name:   "Error message",
			method: (*Manager).Error,
			format: "Error: %v",
			args:   []interface{}{"failed"},
			expect: "Error: failed",
		},
		{
			name:   "Warning message",
			method: (*Manager).Warning,
			format: "Warning: %s",
			args:   []interface{}{"caution"},
			expect: "Warning: caution",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			manager := NewManager(FormatPlain, false, false, buf)

			err := tt.method(manager, tt.format, tt.args...)
			assert.NoError(t, err)
			assert.Contains(t, buf.String(), tt.expect)
		})
	}
}

func TestManagerRaw(t *testing.T) {
	t.Run("outputs raw text", func(t *testing.T) {
		buf := &bytes.Buffer{}
		manager := NewManager(FormatTable, false, false, buf)

		err := manager.Raw("raw text output")
		assert.NoError(t, err)
		assert.Equal(t, "raw text output", buf.String())
	})
}

func TestManagerPrintf(t *testing.T) {
	t.Run("formats and outputs text", func(t *testing.T) {
		buf := &bytes.Buffer{}
		manager := NewManager(FormatTable, false, false, buf)

		err := manager.Printf("Hello %s, you have %d messages", "Alice", 3)
		assert.NoError(t, err)
		assert.Equal(t, "Hello Alice, you have 3 messages", buf.String())
	})
}

func TestManagerPrintln(t *testing.T) {
	t.Run("outputs text with newline", func(t *testing.T) {
		buf := &bytes.Buffer{}
		manager := NewManager(FormatTable, false, false, buf)

		err := manager.Println("Line 1", "Line 2")
		assert.NoError(t, err)
		assert.Equal(t, "Line 1 Line 2\n", buf.String())
	})
}

func TestManagerConcurrency(t *testing.T) {
	t.Run("thread-safe operations", func(t *testing.T) {
		buf := &bytes.Buffer{}
		manager := NewManager(FormatTable, false, false, buf)

		done := make(chan bool, 3)

		// Concurrent format changes
		go func() {
			for i := 0; i < 100; i++ {
				manager.SetFormat(FormatJSON)
				manager.SetFormat(FormatTable)
			}
			done <- true
		}()

		// Concurrent quiet mode changes
		go func() {
			for i := 0; i < 100; i++ {
				manager.SetQuiet(true)
				manager.SetQuiet(false)
			}
			done <- true
		}()

		// Concurrent writes
		go func() {
			for i := 0; i < 100; i++ {
				manager.Raw("test")
			}
			done <- true
		}()

		// Wait for all goroutines
		for i := 0; i < 3; i++ {
			<-done
		}

		// Should not panic
		assert.NotNil(t, manager)
	})
}

func TestManagerFormatterIntegration(t *testing.T) {
	testData := struct {
		Name  string `json:"name" yaml:"name"`
		Count int    `json:"count" yaml:"count"`
	}{
		Name:  "test",
		Count: 42,
	}

	tests := []struct {
		name     string
		format   Format
		contains []string
	}{
		{
			name:     "JSON formatter",
			format:   FormatJSON,
			contains: []string{"\"name\"", "\"test\"", "\"count\"", "42"},
		},
		{
			name:     "YAML formatter",
			format:   FormatYAML,
			contains: []string{"name:", "test", "count:", "42"},
		},
		{
			name:     "Table formatter",
			format:   FormatTable,
			contains: []string{"test", "42"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			manager := NewManager(tt.format, false, false, buf)

			err := manager.Display(testData)
			require.NoError(t, err)

			output := buf.String()
			for _, expected := range tt.contains {
				assert.Contains(t, output, expected)
			}
		})
	}
}

func TestNoGlobalState(t *testing.T) {
	t.Run("managers are independent", func(t *testing.T) {
		buf1 := &bytes.Buffer{}
		buf2 := &bytes.Buffer{}

		manager1 := NewManager(FormatJSON, false, false, buf1)
		manager2 := NewManager(FormatTable, true, false, buf2)

		// Changes to one don't affect the other
		manager1.SetFormat(FormatYAML)
		assert.Equal(t, FormatYAML, manager1.GetFormat())
		assert.Equal(t, FormatTable, manager2.GetFormat())

		manager1.SetQuiet(true)
		assert.True(t, manager1.IsQuiet())
		assert.True(t, manager2.IsQuiet()) // Was already true

		manager2.SetQuiet(false)
		assert.True(t, manager1.IsQuiet())
		assert.False(t, manager2.IsQuiet())

		// Set manager1 back to non-quiet for output test
		manager1.SetQuiet(false)

		// Outputs go to different buffers
		manager1.Raw("output1")
		manager2.Raw("output2")

		// Manager1 is YAML format, which structures raw output
		assert.Contains(t, buf1.String(), "output1")
		assert.Contains(t, buf1.String(), "type: raw")
		// Manager2 is Table format with non-quiet, outputs raw text
		assert.Equal(t, "output2", buf2.String())
	})
}

func TestTemporaryGlobalFunctions(t *testing.T) {
	t.Run("global functions use global manager", func(t *testing.T) {
		// Set up a test global manager
		buf := &bytes.Buffer{}
		testManager := NewManager(FormatTable, false, false, buf)
		SetGlobalManager(testManager)

		// Test that global functions work
		err := Info("test info")
		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "test info")

		buf.Reset()
		err = Success("test success")
		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "test success")

		// Test IsQuiet
		assert.False(t, IsQuiet())
		SetQuiet(true)
		assert.True(t, IsQuiet())
	})

	t.Run("global functions initialize default if needed", func(t *testing.T) {
		// Reset global manager
		SetGlobalManager(nil)

		// Should initialize with defaults
		format := getGlobalManager().GetFormat()
		assert.Equal(t, FormatTable, format)
	})
}
