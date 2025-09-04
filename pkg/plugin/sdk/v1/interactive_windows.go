//go:build windows
// +build windows

package v1

import (
	"os"
	"os/signal"
	"syscall"
)

// setupSignalHandling sets up Windows-specific signal handling
func (s *LocalInteractiveSession) setupSignalHandling(sigCh chan os.Signal) {
	// Windows only supports SIGINT and SIGTERM
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
}

// handleSignal handles Windows-specific signals
func (s *LocalInteractiveSession) handleSignal(sig os.Signal) error {
	// No special handling needed for Windows signals
	return nil
}