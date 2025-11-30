// Package interfaces defines shared interfaces for Glide components.
//
// This package provides interface definitions that allow loose coupling
// between packages, enabling easier testing and dependency injection.
// Implementations are in their respective packages.
//
// # Design Principles
//
// Interfaces are defined here when they need to be shared across packages
// or when defining a contract for dependency injection. This follows the
// Go proverb: "Accept interfaces, return structs."
//
// # Common Patterns
//
// Implement interfaces in concrete types:
//
//	// Interface defined here
//	type Logger interface {
//	    Info(msg string, args ...any)
//	    Error(msg string, args ...any)
//	}
//
//	// Implementation in pkg/logging
//	type SlogLogger struct {
//	    handler slog.Handler
//	}
//
//	func (l *SlogLogger) Info(msg string, args ...any) {
//	    l.handler.Handle(...)
//	}
//
// # Testing
//
// Interfaces enable easy mocking in tests:
//
//	type mockLogger struct {
//	    infoCalls []string
//	}
//
//	func (m *mockLogger) Info(msg string, args ...any) {
//	    m.infoCalls = append(m.infoCalls, msg)
//	}
//
//	func TestWithMock(t *testing.T) {
//	    mock := &mockLogger{}
//	    svc := NewService(mock)
//	    svc.DoSomething()
//	    assert.Contains(t, mock.infoCalls, "operation complete")
//	}
package interfaces
