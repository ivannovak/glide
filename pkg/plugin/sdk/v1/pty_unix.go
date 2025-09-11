//go:build !windows
// +build !windows

package v1

import (
	"os"
	"os/exec"

	"github.com/creack/pty"
)

// startWithPTY starts a command with a PTY on Unix systems
func startWithPTY(cmd *exec.Cmd) (*os.File, error) {
	return pty.Start(cmd)
}

// setPTYSize sets the size of a PTY on Unix systems
func setPTYSize(ptmx *os.File, width, height int) error {
	// Bounds check to prevent integer overflow
	if height < 0 || height > 65535 {
		height = 24 // Default terminal height
	}
	if width < 0 || width > 65535 {
		width = 80 // Default terminal width
	}
	return pty.Setsize(ptmx, &pty.Winsize{
		Rows: uint16(height), // #nosec G115 - bounds checked above
		Cols: uint16(width),  // #nosec G115 - bounds checked above
	})
}
