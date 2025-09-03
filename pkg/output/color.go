package output

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

// ColorConfig holds color configuration
type ColorConfig struct {
	Enabled bool
	Theme   Theme
}

// Theme represents the color theme
type Theme string

const (
	ThemeAuto  Theme = "auto"
	ThemeLight Theme = "light"
	ThemeDark  Theme = "dark"
)

// Colors for different message types
var (
	ColorSuccess = color.New(color.FgGreen)
	ColorError   = color.New(color.FgRed)
	ColorWarning = color.New(color.FgYellow)
	ColorInfo    = color.New(color.FgCyan)
	ColorBold    = color.New(color.Bold)
	ColorFaint   = color.New(color.Faint)
)

// InitColors initializes color settings based on environment
func InitColors() *ColorConfig {
	config := &ColorConfig{
		Enabled: true,
		Theme:   ThemeAuto,
	}

	// Check NO_COLOR environment variable
	if os.Getenv("NO_COLOR") != "" {
		config.Enabled = false
		color.NoColor = true
		return config
	}

	// Check TERM environment variable
	term := os.Getenv("TERM")
	if term == "dumb" || term == "" {
		config.Enabled = false
		color.NoColor = true
		return config
	}

	// Check if output is a terminal
	if !isTerminal() {
		config.Enabled = false
		color.NoColor = true
		return config
	}

	// Detect theme based on terminal background (simplified)
	if os.Getenv("COLORFGBG") != "" {
		// Parse COLORFGBG to determine if dark or light
		parts := strings.Split(os.Getenv("COLORFGBG"), ";")
		if len(parts) >= 2 {
			// This is a simplified check
			// In reality, you'd parse the color values
			config.Theme = ThemeDark
		}
	}

	return config
}

// isTerminal checks if stdout is a terminal
func isTerminal() bool {
	fileInfo, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

// DisableColors disables all color output
func DisableColors() {
	color.NoColor = true
}

// EnableColors enables color output
func EnableColors() {
	color.NoColor = false
}

// Semantic color functions

// SuccessText formats text in success color (green)
func SuccessText(format string, args ...interface{}) string {
	if color.NoColor {
		return fmt.Sprintf(format, args...)
	}
	return ColorSuccess.Sprintf(format, args...)
}

// ErrorText formats text in error color (red)
func ErrorText(format string, args ...interface{}) string {
	if color.NoColor {
		return fmt.Sprintf(format, args...)
	}
	return ColorError.Sprintf(format, args...)
}

// WarningText formats text in warning color (yellow)
func WarningText(format string, args ...interface{}) string {
	if color.NoColor {
		return fmt.Sprintf(format, args...)
	}
	return ColorWarning.Sprintf(format, args...)
}

// InfoText formats text in info color (cyan)
func InfoText(format string, args ...interface{}) string {
	if color.NoColor {
		return fmt.Sprintf(format, args...)
	}
	return ColorInfo.Sprintf(format, args...)
}

// Bold formats text in bold
func Bold(format string, args ...interface{}) string {
	if color.NoColor {
		return fmt.Sprintf(format, args...)
	}
	return ColorBold.Sprintf(format, args...)
}

// Faint formats text in faint/dim style
func Faint(format string, args ...interface{}) string {
	if color.NoColor {
		return fmt.Sprintf(format, args...)
	}
	return ColorFaint.Sprintf(format, args...)
}

// Icons for different message types (with fallbacks for non-unicode terminals)
const (
	IconSuccess = "✓"
	IconError   = "✗"
	IconWarning = "⚠"
	IconInfo    = "ℹ"
	IconBullet  = "•"
	IconArrow   = "→"

	// ASCII fallbacks
	IconSuccessASCII = "[OK]"
	IconErrorASCII   = "[ERROR]"
	IconWarningASCII = "[WARN]"
	IconInfoASCII    = "[INFO]"
	IconBulletASCII  = "*"
	IconArrowASCII   = "->"
)

// GetIcon returns the appropriate icon based on terminal capabilities
func GetIcon(icon string) string {
	// Check if we should use ASCII icons
	if os.Getenv("GLIDE_ASCII_ICONS") != "" || os.Getenv("TERM") == "dumb" {
		switch icon {
		case IconSuccess:
			return IconSuccessASCII
		case IconError:
			return IconErrorASCII
		case IconWarning:
			return IconWarningASCII
		case IconInfo:
			return IconInfoASCII
		case IconBullet:
			return IconBulletASCII
		case IconArrow:
			return IconArrowASCII
		}
	}
	return icon
}
