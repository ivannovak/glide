package output

import (
	"bytes"
	"testing"
)

// BenchmarkManager_Creation benchmarks output manager creation
func BenchmarkManager_Creation(b *testing.B) {
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		writer := &bytes.Buffer{}
		manager := NewManager(FormatTable, false, false, writer)
		_ = manager
	}
}

// BenchmarkManager_Display benchmarks display operations
func BenchmarkManager_Display(b *testing.B) {
	writer := &bytes.Buffer{}
	manager := NewManager(FormatTable, false, false, writer)
	
	testData := map[string]interface{}{
		"name":    "test",
		"version": "1.0.0",
		"active":  true,
		"count":   42,
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		writer.Reset()
		err := manager.Display(testData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkManager_Info benchmarks info message formatting
func BenchmarkManager_Info(b *testing.B) {
	writer := &bytes.Buffer{}
	manager := NewManager(FormatTable, false, false, writer)
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		writer.Reset()
		err := manager.Info("Test info message with %s and %d", "string", 42)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkManager_Success benchmarks success message formatting
func BenchmarkManager_Success(b *testing.B) {
	writer := &bytes.Buffer{}
	manager := NewManager(FormatTable, false, false, writer)
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		writer.Reset()
		err := manager.Success("Operation completed successfully in %v", "1.5s")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkManager_Error benchmarks error message formatting
func BenchmarkManager_Error(b *testing.B) {
	writer := &bytes.Buffer{}
	manager := NewManager(FormatTable, false, false, writer)
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		writer.Reset()
		err := manager.Error("Error occurred: %s (code: %d)", "connection failed", 500)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkManager_Warning benchmarks warning message formatting
func BenchmarkManager_Warning(b *testing.B) {
	writer := &bytes.Buffer{}
	manager := NewManager(FormatTable, false, false, writer)
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		writer.Reset()
		err := manager.Warning("Warning: %s might be outdated", "configuration")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkRegistry_CreateFormatter benchmarks formatter creation from registry
func BenchmarkRegistry_CreateFormatter(b *testing.B) {
	writer := &bytes.Buffer{}
	
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		formatter, err := CreateFormatter(FormatJSON, writer, false, false)
		if err != nil {
			b.Fatal(err)
		}
		_ = formatter
	}
}

// BenchmarkJSONFormatter_Display benchmarks JSON formatting
func BenchmarkJSONFormatter_Display(b *testing.B) {
	writer := &bytes.Buffer{}
	formatter := NewJSONFormatter(writer, false, false)
	
	testData := map[string]interface{}{
		"name":        "benchmark-test",
		"version":     "1.0.0",
		"description": "A test for benchmarking JSON output formatting",
		"features":    []string{"fast", "reliable", "secure"},
		"config": map[string]interface{}{
			"timeout": 30,
			"retries": 3,
			"debug":   false,
		},
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		writer.Reset()
		err := formatter.Display(testData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkYAMLFormatter_Display benchmarks YAML formatting
func BenchmarkYAMLFormatter_Display(b *testing.B) {
	writer := &bytes.Buffer{}
	formatter := NewYAMLFormatter(writer, false, false)
	
	testData := map[string]interface{}{
		"name":        "benchmark-test",
		"version":     "1.0.0",
		"description": "A test for benchmarking YAML output formatting",
		"features":    []string{"readable", "human-friendly", "structured"},
		"config": map[string]interface{}{
			"timeout": 30,
			"retries": 3,
			"debug":   false,
		},
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		writer.Reset()
		err := formatter.Display(testData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkTableFormatter_Display benchmarks table formatting
func BenchmarkTableFormatter_Display(b *testing.B) {
	writer := &bytes.Buffer{}
	formatter := NewTableFormatter(writer, false, false)
	
	testData := []map[string]interface{}{
		{"name": "service-1", "status": "running", "port": 8080, "uptime": "1d 2h"},
		{"name": "service-2", "status": "stopped", "port": 8081, "uptime": "0s"},
		{"name": "service-3", "status": "running", "port": 8082, "uptime": "3h 15m"},
		{"name": "service-4", "status": "running", "port": 8083, "uptime": "5d 12h"},
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		writer.Reset()
		err := formatter.Display(testData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkPlainFormatter_Display benchmarks plain text formatting
func BenchmarkPlainFormatter_Display(b *testing.B) {
	writer := &bytes.Buffer{}
	formatter := NewPlainFormatter(writer, false, false)
	
	testData := map[string]interface{}{
		"name":        "benchmark-test",
		"status":      "active",
		"connections": 1250,
		"uptime":      "2d 4h 30m",
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		writer.Reset()
		err := formatter.Display(testData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkManager_SetFormat benchmarks format switching
func BenchmarkManager_SetFormat(b *testing.B) {
	writer := &bytes.Buffer{}
	manager := NewManager(FormatTable, false, false, writer)
	
	formats := []Format{FormatJSON, FormatYAML, FormatTable, FormatPlain}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		format := formats[i%len(formats)]
		manager.SetFormat(format)
	}
}

// BenchmarkManager_Printf benchmarks printf functionality
func BenchmarkManager_Printf(b *testing.B) {
	writer := &bytes.Buffer{}
	manager := NewManager(FormatPlain, false, false, writer)
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		writer.Reset()
		err := manager.Printf("Benchmark iteration %d with value %s and number %.2f", i, "test", 3.14159)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkManager_Println benchmarks println functionality
func BenchmarkManager_Println(b *testing.B) {
	writer := &bytes.Buffer{}
	manager := NewManager(FormatPlain, false, false, writer)
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		writer.Reset()
		err := manager.Println("Benchmark", "test", "with", "multiple", "arguments", i)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkColors_Enable benchmarks color enabling/disabling
func BenchmarkColors_Enable(b *testing.B) {
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		if i%2 == 0 {
			EnableColors()
		} else {
			DisableColors()
		}
	}
}

// BenchmarkFormat_String benchmarks format string conversion
func BenchmarkFormat_String(b *testing.B) {
	format := FormatJSON
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		str := string(format)
		_ = str
	}
}