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
	return pty.Setsize(ptmx, &pty.Winsize{
		Rows: uint16(height),
		Cols: uint16(width),
	})
}