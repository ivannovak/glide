package errors

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
)

// Handler manages error display and formatting
type Handler struct {
	Writer      io.Writer
	Verbose     bool
	NoColor     bool
	ShowContext bool
}

// DefaultHandler creates a handler with default settings
func DefaultHandler() *Handler {
	return &Handler{
		Writer:      os.Stderr,
		Verbose:     false,
		NoColor:     false,
		ShowContext: false,
	}
}

// Handle processes and displays an error
func (h *Handler) Handle(err error) int {
	if err == nil {
		return 0
	}

	// Check if it's a GlideError
	glideErr, ok := err.(*GlideError)
	if !ok {
		// Handle as generic error
		h.displayGenericError(err)
		return 1
	}

	// Display the error
	h.displayError(glideErr)

	// Display suggestions if available
	if glideErr.HasSuggestions() {
		h.displaySuggestions(glideErr.Suggestions)
	}

	// Display context if verbose mode
	if h.Verbose && len(glideErr.Context) > 0 {
		h.displayContext(glideErr.Context)
	}

	// Return the appropriate exit code
	if glideErr.Code > 0 {
		return glideErr.Code
	}
	return 1
}

// displayError shows the main error message
func (h *Handler) displayError(err *GlideError) {
	icon := h.getErrorIcon(err.Type)
	typeStr := h.getErrorTypeString(err.Type)

	// Build the error message
	var msg strings.Builder

	// Error header
	if h.NoColor {
		fmt.Fprintf(&msg, "%s %s: ", icon, typeStr)
	} else {
		fmt.Fprintf(&msg, "%s %s: ", icon, color.RedString(typeStr))
	}

	// Error message
	msg.WriteString(err.Message)

	// Write to output
	fmt.Fprintln(h.Writer, msg.String())

	// If there's an underlying error and we're in verbose mode, show it
	if h.Verbose && err.Err != nil {
		if h.NoColor {
			fmt.Fprintf(h.Writer, "  Underlying error: %v\n", err.Err)
		} else {
			fmt.Fprintf(h.Writer, "  %s: %v\n", color.HiBlackString("Underlying error"), err.Err)
		}
	}
}

// displayGenericError shows a non-GlideError error
func (h *Handler) displayGenericError(err error) {
	if h.NoColor {
		fmt.Fprintf(h.Writer, "✗ Error: %v\n", err)
	} else {
		fmt.Fprintf(h.Writer, "%s %s: %v\n",
			color.RedString("✗"),
			color.RedString("Error"),
			err)
	}
}

// displaySuggestions shows helpful suggestions
func (h *Handler) displaySuggestions(suggestions []string) {
	if len(suggestions) == 0 {
		return
	}

	fmt.Fprintln(h.Writer)
	if h.NoColor {
		fmt.Fprintln(h.Writer, "Possible solutions:")
	} else {
		fmt.Fprintln(h.Writer, color.YellowString("Possible solutions:"))
	}

	for _, suggestion := range suggestions {
		if h.NoColor {
			fmt.Fprintf(h.Writer, "  • %s\n", suggestion)
		} else {
			// Check if it's a command (starts with common command words)
			if strings.HasPrefix(suggestion, "Run:") ||
				strings.HasPrefix(suggestion, "Check:") ||
				strings.HasPrefix(suggestion, "Fix:") {
				parts := strings.SplitN(suggestion, ":", 2)
				if len(parts) == 2 {
					fmt.Fprintf(h.Writer, "  • %s: %s\n",
						parts[0],
						color.CyanString(strings.TrimSpace(parts[1])))
				} else {
					fmt.Fprintf(h.Writer, "  • %s\n", color.YellowString(suggestion))
				}
			} else {
				fmt.Fprintf(h.Writer, "  • %s\n", color.YellowString(suggestion))
			}
		}
	}
}

// displayContext shows additional context information
func (h *Handler) displayContext(context map[string]string) {
	fmt.Fprintln(h.Writer)
	if h.NoColor {
		fmt.Fprintln(h.Writer, "Context:")
	} else {
		fmt.Fprintln(h.Writer, color.HiBlackString("Context:"))
	}

	for key, value := range context {
		if h.NoColor {
			fmt.Fprintf(h.Writer, "  %s: %s\n", key, value)
		} else {
			fmt.Fprintf(h.Writer, "  %s: %s\n",
				color.HiBlackString(key),
				value)
		}
	}
}

// getErrorIcon returns an appropriate icon for the error type
func (h *Handler) getErrorIcon(errType ErrorType) string {
	switch errType {
	case TypeDocker, TypeContainer:
		return "🐳"
	case TypePermission:
		return "🔒"
	case TypeFileNotFound:
		return "📁"
	case TypeDependency, TypeMissing:
		return "📦"
	case TypeConfig:
		return "⚙️"
	case TypeNetwork, TypeConnection:
		return "🌐"
	case TypeDatabase:
		return "🗄️"
	case TypeMode, TypeWrongMode:
		return "🔄"
	case TypeTimeout:
		return "⏱️"
	case TypeCommand:
		return "💻"
	default:
		return "✗"
	}
}

// getErrorTypeString returns a human-readable error type
func (h *Handler) getErrorTypeString(errType ErrorType) string {
	switch errType {
	case TypeDocker:
		return "Docker Error"
	case TypeContainer:
		return "Container Error"
	case TypePermission:
		return "Permission Error"
	case TypeFileNotFound:
		return "File Not Found"
	case TypeDependency:
		return "Dependency Error"
	case TypeMissing:
		return "Missing Resource"
	case TypeConfig:
		return "Configuration Error"
	case TypeNetwork:
		return "Network Error"
	case TypeConnection:
		return "Connection Error"
	case TypeDatabase:
		return "Database Error"
	case TypeMode:
		return "Mode Error"
	case TypeWrongMode:
		return "Wrong Mode"
	case TypeTimeout:
		return "Timeout"
	case TypeCommand:
		return "Command Error"
	default:
		return "Error"
	}
}

// Print is a convenience function to handle an error with the default handler
func Print(err error) int {
	return DefaultHandler().Handle(err)
}

// PrintVerbose handles an error with verbose output
func PrintVerbose(err error) int {
	handler := DefaultHandler()
	handler.Verbose = true
	return handler.Handle(err)
}

// Exit handles an error and exits with the appropriate code
func Exit(err error) {
	os.Exit(Print(err))
}

// ExitVerbose handles an error verbosely and exits
func ExitVerbose(err error) {
	os.Exit(PrintVerbose(err))
}
