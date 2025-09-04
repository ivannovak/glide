//go:build windows
// +build windows

package v1

import (
	"fmt"
	"os"
	"os/exec"
)

// startWithPTY starts a command with pseudo-terminal support on Windows
// Note: Windows doesn't have true PTY support like Unix, this is a fallback
func startWithPTY(cmd *exec.Cmd) (*os.File, error) {
	// On Windows, we can't use PTY, so we'll use standard pipes
	// This is a simplified implementation that may not support all interactive features
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}
	
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start command: %w", err)
	}
	
	// Return stdin as a file handle (not a true PTY)
	if f, ok := stdin.(*os.File); ok {
		return f, nil
	}
	
	// If stdin is not an *os.File, we need to handle it differently
	// This is a limitation on Windows
	return nil, fmt.Errorf("windows does not support full PTY functionality")
}

// setPTYSize sets the terminal size on Windows (no-op as Windows doesn't have PTY)
func setPTYSize(ptmx *os.File, width, height int) error {
	// Windows doesn't support PTY resize in the same way
	// This is a no-op for now
	return nil
}