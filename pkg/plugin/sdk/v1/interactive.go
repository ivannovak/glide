package v1

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"

	"golang.org/x/term"
)

// InteractiveExecutor handles interactive command execution
type InteractiveExecutor struct {
	host   *Host
	ctx    context.Context
	cancel context.CancelFunc
}

// NewInteractiveExecutor creates a new interactive command executor
func NewInteractiveExecutor(host *Host) *InteractiveExecutor {
	ctx, cancel := context.WithCancel(context.Background())
	return &InteractiveExecutor{
		host:   host,
		ctx:    ctx,
		cancel: cancel,
	}
}

// ExecuteDockerInteractive executes an interactive Docker command
func (e *InteractiveExecutor) ExecuteDockerInteractive(req *DockerRequest) (InteractiveSession, error) {
	session, err := e.host.ExecuteDockerInteractive(e.ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to start interactive Docker session: %w", err)
	}

	return &LocalInteractiveSession{
		session: session,
		ctx:     e.ctx,
	}, nil
}

// ExecuteLocalInteractive executes an interactive local command with PTY
func (e *InteractiveExecutor) ExecuteLocalInteractive(command string, args []string) (InteractiveSession, error) {
	cmd := exec.CommandContext(e.ctx, command, args...)

	// Start the command with a PTY (platform-specific)
	ptmx, err := startWithPTY(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to start command with PTY: %w", err)
	}

	return &PTYSession{
		cmd:  cmd,
		ptmx: ptmx,
		ctx:  e.ctx,
	}, nil
}

// LocalInteractiveSession wraps a remote interactive session for local use
type LocalInteractiveSession struct {
	session InteractiveSession
	ctx     context.Context
}

func (s *LocalInteractiveSession) Send(msg *StreamMessage) error {
	return s.session.Send(msg)
}

func (s *LocalInteractiveSession) Recv() (*StreamMessage, error) {
	return s.session.Recv()
}

func (s *LocalInteractiveSession) Close() error {
	return s.session.Close()
}

// RunInteractiveLoop runs the main interactive loop
func (s *LocalInteractiveSession) RunInteractiveLoop() error {
	// Set up signal handling
	sigCh := make(chan os.Signal, 1)
	s.setupSignalHandling(sigCh)

	// Handle stdin/stdout in separate goroutines
	errCh := make(chan error, 3)

	// Forward stdin to session
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := os.Stdin.Read(buf)
			if err != nil {
				if err != io.EOF {
					errCh <- fmt.Errorf("stdin read error: %w", err)
				}
				return
			}

			if err := s.Send(&StreamMessage{
				Type: StreamMessage_STDIN,
				Data: buf[:n],
			}); err != nil {
				errCh <- fmt.Errorf("failed to send stdin: %w", err)
				return
			}
		}
	}()

	// Forward session output to stdout/stderr
	go func() {
		for {
			msg, err := s.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				errCh <- fmt.Errorf("session recv error: %w", err)
				return
			}

			switch msg.Type {
			case StreamMessage_STDOUT:
				_, _ = os.Stdout.Write(msg.Data)
			case StreamMessage_STDERR:
				_, _ = os.Stderr.Write(msg.Data)
			case StreamMessage_EXIT:
				errCh <- nil
				return
			case StreamMessage_ERROR:
				errCh <- fmt.Errorf("session error: %s", msg.Error)
				return
			}
		}
	}()

	// Handle signals
	go func() {
		for sig := range sigCh {
			// Platform-specific signal handling (e.g., SIGWINCH on Unix)
			if err := s.handleSignal(sig); err != nil {
				// Log error but don't fail the session
				continue
			}

			switch sig {
			case syscall.SIGINT:
				_ = s.Send(&StreamMessage{
					Type:   StreamMessage_SIGNAL,
					Signal: "INT",
				})
			case syscall.SIGTERM:
				_ = s.Send(&StreamMessage{
					Type:   StreamMessage_SIGNAL,
					Signal: "TERM",
				})
				errCh <- nil
				return
			}
		}
	}()

	// Wait for completion or error
	return <-errCh
}

// PTYSession implements InteractiveSession for local PTY processes
type PTYSession struct {
	cmd  *exec.Cmd
	ptmx *os.File
	ctx  context.Context
}

func (s *PTYSession) Send(msg *StreamMessage) error {
	switch msg.Type {
	case StreamMessage_STDIN:
		_, err := s.ptmx.Write(msg.Data)
		return err
	case StreamMessage_SIGNAL:
		return s.handleSignal(msg.Signal)
	case StreamMessage_RESIZE:
		return s.handleResize(int(msg.Width), int(msg.Height))
	}
	return nil
}

func (s *PTYSession) Recv() (*StreamMessage, error) {
	buf := make([]byte, 1024)
	n, err := s.ptmx.Read(buf)
	if err != nil {
		if err == io.EOF {
			// Process has exited
			if s.cmd.ProcessState != nil {
				exitCode := s.cmd.ProcessState.ExitCode()
				// Ensure exit code fits in int32 range
				if exitCode < -2147483648 || exitCode > 2147483647 {
					exitCode = 1 // Default error exit code
				}
				return &StreamMessage{
					Type:     StreamMessage_EXIT,
					ExitCode: int32(exitCode), // #nosec G115 - bounds checked above
				}, nil
			}
			return nil, io.EOF
		}
		return &StreamMessage{
			Type:  StreamMessage_ERROR,
			Error: err.Error(),
		}, nil
	}

	return &StreamMessage{
		Type: StreamMessage_STDOUT,
		Data: buf[:n],
	}, nil
}

func (s *PTYSession) Close() error {
	if s.ptmx != nil {
		_ = s.ptmx.Close()
	}
	if s.cmd != nil && s.cmd.Process != nil {
		return s.cmd.Process.Kill()
	}
	return nil
}

func (s *PTYSession) handleSignal(signal string) error {
	if s.cmd == nil || s.cmd.Process == nil {
		return nil
	}

	switch signal {
	case "INT":
		return s.cmd.Process.Signal(syscall.SIGINT)
	case "TERM":
		return s.cmd.Process.Signal(syscall.SIGTERM)
	case "KILL":
		return s.cmd.Process.Kill()
	}
	return nil
}

func (s *PTYSession) handleResize(width, height int) error {
	return setPTYSize(s.ptmx, width, height)
}

// InteractiveCommand provides a high-level interface for interactive commands
type InteractiveCommand struct {
	executor *InteractiveExecutor
}

// NewInteractiveCommand creates a new interactive command helper
func NewInteractiveCommand(host *Host) *InteractiveCommand {
	return &InteractiveCommand{
		executor: NewInteractiveExecutor(host),
	}
}

// RunDockerExec runs an interactive docker exec command
func (ic *InteractiveCommand) RunDockerExec(container, command string, args []string) error {
	dockerArgs := []string{"exec", "-it", container, command}
	dockerArgs = append(dockerArgs, args...)

	session, err := ic.executor.ExecuteDockerInteractive(&DockerRequest{
		Operation:   "exec",
		Args:        dockerArgs,
		Interactive: true,
		TTY:         true,
	})
	if err != nil {
		return err
	}
	defer session.Close()

	if localSession, ok := session.(*LocalInteractiveSession); ok {
		return localSession.RunInteractiveLoop()
	}

	return fmt.Errorf("unsupported session type")
}

// RunDockerCompose runs an interactive docker-compose command
func (ic *InteractiveCommand) RunDockerCompose(args []string, workDir string) error {
	session, err := ic.executor.ExecuteDockerInteractive(&DockerRequest{
		Operation:   "compose",
		Args:        args,
		WorkDir:     workDir,
		Interactive: true,
		TTY:         true,
	})
	if err != nil {
		return err
	}
	defer session.Close()

	if localSession, ok := session.(*LocalInteractiveSession); ok {
		return localSession.RunInteractiveLoop()
	}

	return fmt.Errorf("unsupported session type")
}

// RunLocal runs an interactive local command
func (ic *InteractiveCommand) RunLocal(command string, args []string) error {
	session, err := ic.executor.ExecuteLocalInteractive(command, args)
	if err != nil {
		return err
	}
	defer session.Close()

	if ptySession, ok := session.(*PTYSession); ok {
		// Set up raw mode for terminal
		oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
		if err != nil {
			return fmt.Errorf("failed to set raw mode: %w", err)
		}
		defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }()

		// Copy between PTY and terminal
		errCh := make(chan error, 2)

		go func() {
			_, err := io.Copy(ptySession.ptmx, os.Stdin)
			errCh <- err
		}()

		go func() {
			_, err := io.Copy(os.Stdout, ptySession.ptmx)
			errCh <- err
		}()

		// Wait for process to exit or error
		go func() {
			errCh <- ptySession.cmd.Wait()
		}()

		return <-errCh
	}

	return fmt.Errorf("unsupported session type")
}
