// Package mocks provides mock implementations for testing.
//
// This package contains mock implementations of interfaces used
// throughout Glide, enabling isolated unit testing without external
// dependencies.
//
// # Usage in Tests
//
// Create mock instances for testing:
//
//	func TestWithMocks(t *testing.T) {
//	    mockLogger := mocks.NewMockLogger()
//	    mockConfig := mocks.NewMockConfig()
//
//	    service := NewService(mockLogger, mockConfig)
//	    err := service.DoSomething()
//
//	    assert.NoError(t, err)
//	    assert.Contains(t, mockLogger.InfoCalls(), "operation complete")
//	}
//
// # Recording Calls
//
// Mocks record method calls for verification:
//
//	mock := mocks.NewMockExecutor()
//	service.Execute()
//
//	calls := mock.Calls()
//	assert.Equal(t, 1, len(calls))
//	assert.Equal(t, "Execute", calls[0].Method)
//
// # Setting Return Values
//
// Configure mock behavior:
//
//	mock := mocks.NewMockLoader()
//	mock.OnLoad(func() (*Config, error) {
//	    return &Config{Debug: true}, nil
//	})
//
// # Error Simulation
//
// Simulate errors for testing error paths:
//
//	mock := mocks.NewMockExecutor()
//	mock.SetError(errors.New("command failed"))
//
//	err := service.Execute()
//	assert.Error(t, err)
package mocks
