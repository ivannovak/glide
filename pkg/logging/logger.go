package logging

import (
	"context"
	"log/slog"
	"runtime"
	"sync"
	"time"
)

// Logger provides structured logging functionality wrapping log/slog
type Logger struct {
	handler slog.Handler
	level   *slog.LevelVar
}

var (
	// defaultLogger is the global logger instance
	defaultLogger *Logger
	once          sync.Once
)

// New creates a new Logger with the specified configuration
func New(config *Config) *Logger {
	levelVar := &slog.LevelVar{}
	levelVar.Set(config.Level)

	var handler slog.Handler
	if config.Format == FormatJSON {
		handler = slog.NewJSONHandler(config.Output, &slog.HandlerOptions{
			Level:     levelVar,
			AddSource: config.AddSource,
		})
	} else {
		handler = slog.NewTextHandler(config.Output, &slog.HandlerOptions{
			Level:     levelVar,
			AddSource: config.AddSource,
		})
	}

	return &Logger{
		handler: handler,
		level:   levelVar,
	}
}

// Default returns the default global logger
func Default() *Logger {
	once.Do(func() {
		defaultLogger = New(DefaultConfig())
	})
	return defaultLogger
}

// SetDefault sets the global default logger
func SetDefault(logger *Logger) {
	once.Do(func() {}) // Ensure once is marked as done
	defaultLogger = logger
}

// SetLevel changes the minimum log level
func (l *Logger) SetLevel(level slog.Level) {
	l.level.Set(level)
}

// Debug logs a debug-level message
func (l *Logger) Debug(msg string, args ...any) {
	l.log(context.Background(), slog.LevelDebug, msg, args...)
}

// Info logs an info-level message
func (l *Logger) Info(msg string, args ...any) {
	l.log(context.Background(), slog.LevelInfo, msg, args...)
}

// Warn logs a warning-level message
func (l *Logger) Warn(msg string, args ...any) {
	l.log(context.Background(), slog.LevelWarn, msg, args...)
}

// Error logs an error-level message
func (l *Logger) Error(msg string, args ...any) {
	l.log(context.Background(), slog.LevelError, msg, args...)
}

// DebugContext logs a debug-level message with context
func (l *Logger) DebugContext(ctx context.Context, msg string, args ...any) {
	l.log(ctx, slog.LevelDebug, msg, args...)
}

// InfoContext logs an info-level message with context
func (l *Logger) InfoContext(ctx context.Context, msg string, args ...any) {
	l.log(ctx, slog.LevelInfo, msg, args...)
}

// WarnContext logs a warning-level message with context
func (l *Logger) WarnContext(ctx context.Context, msg string, args ...any) {
	l.log(ctx, slog.LevelWarn, msg, args...)
}

// ErrorContext logs an error-level message with context
func (l *Logger) ErrorContext(ctx context.Context, msg string, args ...any) {
	l.log(ctx, slog.LevelError, msg, args...)
}

// With returns a new Logger with additional attributes
func (l *Logger) With(args ...any) *Logger {
	return &Logger{
		handler: l.handler.WithAttrs(argsToAttrs(args)),
		level:   l.level,
	}
}

// WithGroup returns a new Logger with a group name
func (l *Logger) WithGroup(name string) *Logger {
	return &Logger{
		handler: l.handler.WithGroup(name),
		level:   l.level,
	}
}

func (l *Logger) log(ctx context.Context, level slog.Level, msg string, args ...any) {
	if !l.handler.Enabled(ctx, level) {
		return
	}
	var pcs [1]uintptr
	// Skip: runtime.Callers, this function, the public method
	runtime_Callers(3, pcs[:])
	r := slog.NewRecord(timeNow(), level, msg, pcs[0])
	r.Add(args...)
	// Safe to ignore: slog.Handler.Handle rarely fails, and if it does, we can't log the error
	// (infinite recursion). Handler implementations are expected to not error on normal use.
	_ = l.handler.Handle(ctx, r)
}

// Helper functions for converting args to attributes
func argsToAttrs(args []any) []slog.Attr {
	attrs := make([]slog.Attr, 0, len(args)/2)
	for i := 0; i < len(args)-1; i += 2 {
		key, ok := args[i].(string)
		if !ok {
			continue
		}
		attrs = append(attrs, slog.Any(key, args[i+1]))
	}
	return attrs
}

// Field helper functions for common types
func String(key, value string) any {
	return slog.String(key, value)
}

func Int(key string, value int) any {
	return slog.Int(key, value)
}

func Int64(key string, value int64) any {
	return slog.Int64(key, value)
}

func Bool(key string, value bool) any {
	return slog.Bool(key, value)
}

func Err(err error) any {
	return slog.Any("error", err)
}

func Duration(key string, value interface{}) any {
	return slog.Any(key, value)
}

// Global convenience functions that use the default logger
func Debug(msg string, args ...any) {
	Default().Debug(msg, args...)
}

func Info(msg string, args ...any) {
	Default().Info(msg, args...)
}

func Warn(msg string, args ...any) {
	Default().Warn(msg, args...)
}

func Error(msg string, args ...any) {
	Default().Error(msg, args...)
}

func DebugContext(ctx context.Context, msg string, args ...any) {
	Default().DebugContext(ctx, msg, args...)
}

func InfoContext(ctx context.Context, msg string, args ...any) {
	Default().InfoContext(ctx, msg, args...)
}

func WarnContext(ctx context.Context, msg string, args ...any) {
	Default().WarnContext(ctx, msg, args...)
}

func ErrorContext(ctx context.Context, msg string, args ...any) {
	Default().ErrorContext(ctx, msg, args...)
}

// SetLevel sets the level on the default logger
func SetLevel(level slog.Level) {
	Default().SetLevel(level)
}

// runtime_Callers is a wrapper for runtime.Callers
func runtime_Callers(skip int, pc []uintptr) int {
	return runtime.Callers(skip, pc)
}

// timeNow returns the current time
func timeNow() time.Time {
	return time.Now()
}
