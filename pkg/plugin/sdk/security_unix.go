//go:build !windows

package sdk

import (
	"fmt"
	"os"
	"syscall"
)

// validateOwnership checks file ownership (Unix-specific)
func (sv *SecurityValidator) validateOwnership(info os.FileInfo) error {
	// Get the underlying syscall.Stat_t
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		// If we can't get syscall info, skip this check
		return nil
	}

	// Get current user's UID
	currentUID := uint32(os.Getuid())

	// Plugin must be owned by current user or root (UID 0)
	if stat.Uid != currentUID && stat.Uid != 0 {
		return fmt.Errorf("plugin file is owned by UID %d, expected current user (%d) or root (0)", stat.Uid, currentUID)
	}

	return nil
}
