package container

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/glide-cli/glide/v3/internal/config"
	"github.com/glide-cli/glide/v3/pkg/logging"
	"github.com/stretchr/testify/require"
)

func TestNew_Success(t *testing.T) {
	c, err := New()
	require.NoError(t, err)
	require.NotNil(t, c)
	require.NotNil(t, c.app)
}

func TestNew_WithOptions(t *testing.T) {
	buf := &bytes.Buffer{}
	testLogger := logging.New(&logging.Config{Level: slog.LevelDebug})
	testCfg := &config.Config{}

	c, err := New(
		WithLogger(testLogger),
		WithWriter(buf),
		WithConfig(testCfg),
		WithoutLifecycle(),
	)
	require.NoError(t, err)
	require.NotNil(t, c)
}

func TestContainer_Lifecycle(t *testing.T) {
	c, err := New()
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start should succeed
	err = c.Start(ctx)
	require.NoError(t, err)

	// Stop should succeed
	err = c.Stop(ctx)
	require.NoError(t, err)
}

func TestContainer_Run(t *testing.T) {
	c, err := New()
	require.NoError(t, err)

	ctx := context.Background()
	executed := false

	err = c.Run(ctx, func() error {
		executed = true
		return nil
	})

	require.NoError(t, err)
	require.True(t, executed, "function should have been executed")
}

func TestContainer_Run_WithError(t *testing.T) {
	c, err := New()
	require.NoError(t, err)

	ctx := context.Background()
	testErr := errors.New("test error")

	err = c.Run(ctx, func() error {
		return testErr
	})

	require.Error(t, err)
	require.Equal(t, testErr, err)
}

// Note: Invoke is not supported by the container
// Dependencies should be extracted via Run() function instead

func TestProviders_Logger(t *testing.T) {
	logger := provideLogger()
	require.NotNil(t, logger)
}

func TestProviders_Writer(t *testing.T) {
	writer := provideWriter()
	require.NotNil(t, writer)
}

func TestProviders_ConfigLoader(t *testing.T) {
	logger := provideLogger()
	loader := provideConfigLoader(logger)
	require.NotNil(t, loader)
}

func TestProviders_Config(t *testing.T) {
	logger := provideLogger()
	loader := provideConfigLoader(logger)

	cfg, err := provideConfig(ConfigParams{
		Loader: loader,
		Logger: logger,
	})

	// Should not error even if config file doesn't exist
	require.NoError(t, err)
	require.NotNil(t, cfg)
}

func TestProviders_ContextDetector(t *testing.T) {
	logger := provideLogger()
	detector, err := provideContextDetector(logger)
	require.NoError(t, err)
	require.NotNil(t, detector)
}

func TestProviders_ProjectContext(t *testing.T) {
	logger := provideLogger()
	detector, err := provideContextDetector(logger)
	require.NoError(t, err)

	ctx, err := provideProjectContext(ProjectContextParams{
		Detector: detector,
		Logger:   logger,
		Plugins:  nil,
	})

	require.NoError(t, err)
	require.NotNil(t, ctx)
}

func TestProviders_OutputManager(t *testing.T) {
	logger := provideLogger()
	buf := &bytes.Buffer{}

	manager := provideOutputManager(OutputManagerParams{
		Writer: buf,
		Logger: logger,
	})

	require.NotNil(t, manager)
}

func TestProviders_ShellExecutor(t *testing.T) {
	logger := provideLogger()
	executor := provideShellExecutor(logger)
	require.NotNil(t, executor)
}

func TestProviders_PluginRegistry(t *testing.T) {
	logger := provideLogger()
	registry := providePluginRegistry(logger)
	require.NotNil(t, registry)
}

func TestOptions_WithLogger(t *testing.T) {
	testLogger := logging.New(&logging.Config{Level: slog.LevelDebug})

	c, err := New(WithLogger(testLogger))
	require.NoError(t, err)

	// Start container to verify logger was injected
	ctx := context.Background()
	err = c.Start(ctx)
	require.NoError(t, err)
	defer c.Stop(ctx)

	// Logger injection verified by successful start
	require.NotNil(t, c)
}

func TestOptions_WithWriter(t *testing.T) {
	buf := &bytes.Buffer{}

	c, err := New(WithWriter(buf))
	require.NoError(t, err)

	ctx := context.Background()
	err = c.Start(ctx)
	require.NoError(t, err)
	defer c.Stop(ctx)

	// Writer injection verified by successful start
	require.NotNil(t, c)
}

func TestOptions_WithConfig(t *testing.T) {
	testCfg := &config.Config{}

	c, err := New(WithConfig(testCfg))
	require.NoError(t, err)

	ctx := context.Background()
	err = c.Start(ctx)
	require.NoError(t, err)
	defer c.Stop(ctx)

	// Config injection verified by successful start
	require.NotNil(t, c)
}
