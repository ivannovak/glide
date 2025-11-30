//go:build windows

package sdk

import (
	"os"
)

// validateOwnership checks file ownership (Windows stub - ownership checks not applicable)
func (sv *SecurityValidator) validateOwnership(info os.FileInfo) error {
	// Windows doesn't have Unix-style ownership, skip this check
	return nil
}
