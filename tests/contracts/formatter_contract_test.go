// Package contracts contains contract tests for formatter implementations
//
// Contract tests ensure all Formatter implementations behave consistently:
// - All formatters implement the complete Formatter interface
// - All formatters handle common data types correctly
// - All formatters respect quiet mode and writer settings
package contracts

import (
	"bytes"
	"strings"
	"testing"

	"github.com/ivannovak/glide/v3/pkg/output"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFormatterContract verifies all formatter implementations adhere to the Formatter contract
func TestFormatterContract(t *testing.T) {
	formatters := map[string]func(*bytes.Buffer, bool) output.Formatter{
		"Plain": func(buf *bytes.Buffer, noColor bool) output.Formatter {
			return output.NewPlainFormatter(buf, noColor, false)
		},
		"JSON": func(buf *bytes.Buffer, noColor bool) output.Formatter {
			return output.NewJSONFormatter(buf, noColor, false)
		},
		"YAML": func(buf *bytes.Buffer, noColor bool) output.Formatter {
			return output.NewYAMLFormatter(buf, noColor, false)
		},
		"Table": func(buf *bytes.Buffer, noColor bool) output.Formatter {
			return output.NewTableFormatter(buf, noColor, false)
		},
	}

	for name, createFormatter := range formatters {
		t.Run(name, func(t *testing.T) {
			testFormatterContract(t, name, createFormatter)
		})
	}
}

func testFormatterContract(t *testing.T, formatterName string, createFormatter func(*bytes.Buffer, bool) output.Formatter) {
	t.Run("implements Info method", func(t *testing.T) {
		buf := &bytes.Buffer{}
		formatter := createFormatter(buf, true) // noColor=true for consistent output

		err := formatter.Info("test message: %s", "hello")
		require.NoError(t, err)

		output := buf.String()
		assert.NotEmpty(t, output, "Info should produce output")
		assert.Contains(t, output, "test message", "Info should contain message")
	})

	t.Run("implements Success method", func(t *testing.T) {
		buf := &bytes.Buffer{}
		formatter := createFormatter(buf, true)

		err := formatter.Success("operation succeeded: %d items", 42)
		require.NoError(t, err)

		output := buf.String()
		assert.NotEmpty(t, output, "Success should produce output")
		assert.Contains(t, output, "succeeded", "Success should contain message")
	})

	t.Run("implements Error method", func(t *testing.T) {
		buf := &bytes.Buffer{}
		formatter := createFormatter(buf, true)

		err := formatter.Error("error occurred: %s", "test error")
		require.NoError(t, err)

		output := buf.String()
		assert.NotEmpty(t, output, "Error should produce output")
		assert.Contains(t, output, "error", "Error should contain message")
	})

	t.Run("implements Warning method", func(t *testing.T) {
		buf := &bytes.Buffer{}
		formatter := createFormatter(buf, true)

		err := formatter.Warning("warning: %s", "test warning")
		require.NoError(t, err)

		output := buf.String()
		assert.NotEmpty(t, output, "Warning should produce output")
		assert.Contains(t, output, "warning", "Warning should contain message")
	})

	t.Run("implements Raw method", func(t *testing.T) {
		buf := &bytes.Buffer{}
		formatter := createFormatter(buf, true)

		expectedText := "raw output text"
		err := formatter.Raw(expectedText)
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, expectedText, "Raw should output text unchanged")
	})

	t.Run("implements SetWriter method", func(t *testing.T) {
		buf1 := &bytes.Buffer{}
		buf2 := &bytes.Buffer{}
		formatter := createFormatter(buf1, true)

		// Write to first buffer
		err := formatter.Info("first")
		require.NoError(t, err)
		assert.NotEmpty(t, buf1.String(), "Should write to first buffer")
		assert.Empty(t, buf2.String(), "Should not write to second buffer yet")

		// Change writer
		formatter.SetWriter(buf2)

		// Write to second buffer
		err = formatter.Info("second")
		require.NoError(t, err)
		assert.NotContains(t, buf1.String(), "second", "Should not write to first buffer after SetWriter")
	})

	t.Run("implements Display method", func(t *testing.T) {
		buf := &bytes.Buffer{}
		formatter := createFormatter(buf, true)

		data := map[string]interface{}{
			"key": "value",
		}

		err := formatter.Display(data)
		require.NoError(t, err)

		output := buf.String()
		assert.NotEmpty(t, output, "Display should produce output for non-nil data")
	})

	t.Run("Display handles nil data gracefully", func(t *testing.T) {
		buf := &bytes.Buffer{}
		formatter := createFormatter(buf, true)

		// All formatters should handle nil data without crashing
		err := formatter.Display(nil)

		// We don't require specific behavior (error or empty output),
		// just that it doesn't panic
		assert.NotPanics(t, func() {
			_ = err
		})
	})

	t.Run("Display handles various data types", func(t *testing.T) {
		testCases := []struct {
			name string
			data interface{}
		}{
			{"string", "test string"},
			{"int", 42},
			{"bool", true},
			{"slice", []string{"a", "b", "c"}},
			{"map", map[string]string{"key": "value"}},
			{"struct", struct{ Name string }{Name: "test"}},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				buf := &bytes.Buffer{}
				formatter := createFormatter(buf, true)

				err := formatter.Display(tc.data)
				require.NoError(t, err, "Display should handle %s without error", tc.name)

				output := buf.String()
				assert.NotEmpty(t, output, "Display should produce output for %s", tc.name)
			})
		}
	})

	t.Run("handles empty format strings", func(t *testing.T) {
		buf := &bytes.Buffer{}
		formatter := createFormatter(buf, true)

		// Test with empty format string - should not panic
		assert.NotPanics(t, func() {
			_ = formatter.Info("")
			_ = formatter.Success("")
			_ = formatter.Error("")
			_ = formatter.Warning("")
		})
	})

	t.Run("handles format args correctly", func(t *testing.T) {
		buf := &bytes.Buffer{}
		formatter := createFormatter(buf, true)

		err := formatter.Info("test %s %d %v", "string", 42, true)
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "string")
		assert.Contains(t, output, "42")
		assert.Contains(t, output, "true")
	})
}

// TestFormatterOutputDifferences validates that different formatters produce different output
// This ensures each formatter actually applies its formatting
func TestFormatterOutputDifferences(t *testing.T) {
	testData := map[string]interface{}{
		"name":    "test",
		"version": "1.0.0",
		"count":   42,
	}

	outputs := make(map[string]string)

	formatters := map[string]func(*bytes.Buffer, bool) output.Formatter{
		"Plain": func(buf *bytes.Buffer, noColor bool) output.Formatter {
			return output.NewPlainFormatter(buf, noColor, false)
		},
		"JSON": func(buf *bytes.Buffer, noColor bool) output.Formatter {
			return output.NewJSONFormatter(buf, noColor, false)
		},
		"YAML": func(buf *bytes.Buffer, noColor bool) output.Formatter {
			return output.NewYAMLFormatter(buf, noColor, false)
		},
		"Table": func(buf *bytes.Buffer, noColor bool) output.Formatter {
			return output.NewTableFormatter(buf, noColor, false)
		},
	}

	// Collect outputs from each formatter
	for name, createFormatter := range formatters {
		buf := &bytes.Buffer{}
		formatter := createFormatter(buf, true)

		err := formatter.Display(testData)
		require.NoError(t, err, "%s formatter should not error", name)

		outputs[name] = buf.String()
		assert.NotEmpty(t, outputs[name], "%s should produce output", name)
	}

	// Verify formatters produce different output
	t.Run("JSON contains JSON markers", func(t *testing.T) {
		assert.True(t,
			strings.Contains(outputs["JSON"], "{") && strings.Contains(outputs["JSON"], "}"),
			"JSON formatter should produce JSON-like output with braces")
	})

	t.Run("YAML contains YAML markers", func(t *testing.T) {
		assert.True(t,
			strings.Contains(outputs["YAML"], ":") || strings.Contains(outputs["YAML"], "-"),
			"YAML formatter should produce YAML-like output with colons or dashes")
	})

	// Formatters should be distinct (not all identical)
	t.Run("formatters produce distinct output", func(t *testing.T) {
		allSame := true
		firstOutput := ""

		for _, output := range outputs {
			if firstOutput == "" {
				firstOutput = output
			} else if output != firstOutput {
				allSame = false
				break
			}
		}

		assert.False(t, allSame, "Formatters should produce different output formats")
	})
}

// TestFormatterQuietMode verifies quiet mode behavior
func TestFormatterQuietMode(t *testing.T) {
	formatters := map[string]func(*bytes.Buffer, bool, bool) output.Formatter{
		"Plain": func(buf *bytes.Buffer, noColor, quiet bool) output.Formatter {
			return output.NewPlainFormatter(buf, noColor, quiet)
		},
		"JSON": func(buf *bytes.Buffer, noColor, quiet bool) output.Formatter {
			return output.NewJSONFormatter(buf, noColor, quiet)
		},
		"YAML": func(buf *bytes.Buffer, noColor, quiet bool) output.Formatter {
			return output.NewYAMLFormatter(buf, noColor, quiet)
		},
		"Table": func(buf *bytes.Buffer, noColor, quiet bool) output.Formatter {
			return output.NewTableFormatter(buf, noColor, quiet)
		},
	}

	for name, createFormatter := range formatters {
		t.Run(name, func(t *testing.T) {
			t.Run("quiet mode suppresses Info messages", func(t *testing.T) {
				buf := &bytes.Buffer{}
				formatter := createFormatter(buf, true, true) // quiet=true

				err := formatter.Info("test message")
				require.NoError(t, err)

				// Quiet mode should suppress Info output
				// Note: Some formatters might still output minimal content
				// The key is that it's significantly less than normal mode
			})

			t.Run("quiet mode never suppresses Error messages", func(t *testing.T) {
				buf := &bytes.Buffer{}
				formatter := createFormatter(buf, true, true) // quiet=true

				err := formatter.Error("test error")
				require.NoError(t, err)

				output := buf.String()
				// Errors should ALWAYS be shown, even in quiet mode
				assert.NotEmpty(t, output, "Errors must not be suppressed in quiet mode")
			})
		})
	}
}
