package winsandbox

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

// Sandbox represents a Windows Sandbox instance
type Sandbox struct {
	config     *Config
	configPath string
	cmd        *exec.Cmd
	ctx        context.Context
	cancel     context.CancelFunc
	started    bool
}

// New creates a new Sandbox instance with the given configuration
func New(config *Config) (*Sandbox, error) {
	if runtime.GOOS != "windows" {
		return nil, fmt.Errorf("Windows Sandbox is only supported on Windows")
	}

	// Check if Windows Sandbox is installed
	installed, err := IsInstalled()
	if err != nil {
		return nil, fmt.Errorf("failed to check if Windows Sandbox is installed: %w", err)
	}
	if !installed {
		return nil, ErrNotInstalled{}
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Sandbox{
		config:  config,
		ctx:     ctx,
		cancel:  cancel,
		started: false,
	}, nil
}

// Start launches the Windows Sandbox with the configured settings
func (s *Sandbox) Start() error {
	if s.started {
		return fmt.Errorf("sandbox is already running")
	}

	// Create temporary config file
	tempDir := os.TempDir()
	configPath := filepath.Join(tempDir, fmt.Sprintf("sandbox_%d.wsb", time.Now().UnixNano()))

	if err := s.config.WriteToFile(configPath); err != nil {
		return fmt.Errorf("failed to write configuration file: %w", err)
	}
	s.configPath = configPath

	// Get path to WindowsSandbox.exe
	windir := os.Getenv("windir")
	if windir == "" {
		windir = "C:\\Windows"
	}
	sandboxExe := filepath.Join(windir, "System32", "WindowsSandbox.exe")

	// Check if WindowsSandbox.exe exists
	if _, err := os.Stat(sandboxExe); os.IsNotExist(err) {
		return ErrSandboxFailed{
			Operation: "start",
			Message:   "WindowsSandbox.exe not found. Ensure Windows Sandbox is installed correctly.",
		}
	}

	// Launch Windows Sandbox
	s.cmd = exec.CommandContext(s.ctx, sandboxExe, configPath)

	var stderr bytes.Buffer
	s.cmd.Stderr = &stderr

	if err := s.cmd.Start(); err != nil {
		return ErrSandboxFailed{
			Operation: "start",
			Message:   fmt.Sprintf("failed to launch: %v\nStderr: %s", err, stderr.String()),
		}
	}

	s.started = true

	// Wait a moment for the sandbox to initialize
	time.Sleep(2 * time.Second)

	// Check if process is still running
	if s.cmd.Process != nil {
		// Process exists, assume it's running successfully
		// We can't reliably check if it failed immediately without waiting for it to exit
		go func() {
			s.cmd.Wait()
			s.started = false
		}()
	}

	return nil
}

// Execute runs a command inside the sandbox
// Note: This is a simplified implementation. True command execution inside sandbox
// requires more complex inter-process communication or using LogonCommand in config
func (s *Sandbox) Execute(command string) (stdout, stderr string, err error) {
	if !s.started {
		return "", "", fmt.Errorf("sandbox is not running")
	}

	// This is a limitation of Windows Sandbox - we cannot easily execute commands
	// inside a running sandbox without advanced techniques.
	// The recommended approach is to use LogonCommand in the configuration.
	return "", "", fmt.Errorf("direct command execution in running sandbox is not supported. Use LogonCommand in configuration instead")
}

// Stop gracefully shuts down the Windows Sandbox
func (s *Sandbox) Stop() error {
	if !s.started {
		return nil
	}

	// Cancel the context to signal shutdown
	if s.cancel != nil {
		s.cancel()
	}

	// Wait for process to exit (with timeout)
	if s.cmd != nil && s.cmd.Process != nil {
		done := make(chan error, 1)
		go func() {
			done <- s.cmd.Wait()
		}()

		select {
		case <-done:
			// Process exited
		case <-time.After(10 * time.Second):
			// Force kill if it doesn't exit gracefully
			if s.cmd.Process != nil {
				s.cmd.Process.Kill()
			}
		}
	}

	s.started = false

	// Clean up temporary config file
	if s.configPath != "" {
		os.Remove(s.configPath)
		s.configPath = ""
	}

	return nil
}

// IsRunning returns true if the sandbox is currently running
func (s *Sandbox) IsRunning() bool {
	if !s.started {
		return false
	}

	// Check if process is still alive
	if s.cmd == nil || s.cmd.Process == nil {
		return false
	}

	// Try to find the process (this is platform-specific)
	// For simplicity, we'll just return the started flag
	return s.started
}

// Wait waits for the sandbox to exit and returns any error
func (s *Sandbox) Wait() error {
	if !s.started || s.cmd == nil {
		return fmt.Errorf("sandbox is not running")
	}

	return s.cmd.Wait()
}

// GetConfigPath returns the path to the temporary configuration file
func (s *Sandbox) GetConfigPath() string {
	return s.configPath
}
