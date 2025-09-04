//go:build !windows
// +build !windows

package v1

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/creack/pty"
)

// setupSignalHandling sets up Unix-specific signal handling
func (s *LocalInteractiveSession) setupSignalHandling(sigCh chan os.Signal) {
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGWINCH)
}

// handleSignal handles Unix-specific signals
func (s *LocalInteractiveSession) handleSignal(sig os.Signal) error {
	if sig == syscall.SIGWINCH {
		// Handle terminal resize
		if ws, err := pty.GetsizeFull(os.Stdin); err == nil {
			return s.Send(&StreamMessage{
				Type:   StreamMessage_RESIZE,
				Width:  int32(ws.Cols),
				Height: int32(ws.Rows),
			})
		}
	}
	return nil
}