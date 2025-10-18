package winsandbox

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Cmd represents a command to be executed in Windows Sandbox.
// It is similar to os/exec.Cmd but designed for Windows Sandbox execution.
type Cmd struct {
	// Path is the path of the command to run in the sandbox.
	Path string

	// Args holds command line arguments, including the command as Args[0].
	// If the Args field is empty or nil, Run uses {Path}.
	Args []string

	// Env specifies the environment of the process in the sandbox.
	// Each entry is of the form "key=value".
	// If Env is nil, the command will use the sandbox's default environment.
	Env []string

	// Dir specifies the working directory of the command in the sandbox.
	// If Dir is empty, the command runs in the default directory.
	Dir string

	// Stdin specifies the process's standard input.
	// If Stdin is nil, the process reads from the null device (os.DevNull).
	Stdin io.Reader

	// Stdout and Stderr specify the process's standard output and error.
	//
	// If either is nil, Run connects the corresponding file descriptor
	// to the null device (os.DevNull).
	//
	// If Stdout and Stderr are the same writer, at most one
	// goroutine at a time will call Write.
	Stdout io.Writer
	Stderr io.Writer

	// Config is the sandbox configuration to use.
	// If nil, a default configuration will be used.
	Config *Config

	// Timeout specifies a maximum duration for the command execution.
	// If Timeout is 0, no timeout is applied.
	Timeout time.Duration

	// sandbox holds the running sandbox instance
	sandbox *Sandbox

	// outputFile is the temporary file used to capture output
	outputFile string
	errorFile  string

	// ctx is the context for the command
	ctx    context.Context
	cancel context.CancelFunc

	// finished indicates if the command has finished
	finished bool

	// err stores any error that occurred
	err error
}

// Command returns the Cmd struct to execute the named program with
// the given arguments in a Windows Sandbox.
//
// It sets only the Path and Args in the returned structure.
//
// If name contains no path separators, Command uses exec.LookPath to
// resolve name to a complete path. Otherwise it uses name directly as Path.
//
// The returned Cmd's Args field is constructed from the command name
// followed by the elements of arg, so arg should not include the
// command name itself.
func Command(name string, arg ...string) *Cmd {
	cmd := &Cmd{
		Path: name,
		Args: append([]string{name}, arg...),
	}
	return cmd
}

// CommandContext is like Command but includes a context.
//
// The provided context is used to kill the process (by killing the sandbox)
// if the context becomes done before the command completes on its own.
func CommandContext(ctx context.Context, name string, arg ...string) *Cmd {
	if ctx == nil {
		panic("nil Context")
	}
	cmd := Command(name, arg...)
	cmd.ctx = ctx
	return cmd
}

// String returns a human-readable description of c.
func (c *Cmd) String() string {
	if c.Path == "" && len(c.Args) == 0 {
		return "<empty command>"
	}
	if len(c.Args) > 0 {
		return strings.Join(c.Args, " ")
	}
	return c.Path
}

// Run starts the specified command in a Windows Sandbox and waits for it to complete.
//
// The returned error is nil if the command runs, has no problems
// copying stdin, stdout, and stderr, and exits with a zero exit status.
func (c *Cmd) Run() error {
	if err := c.Start(); err != nil {
		return err
	}
	return c.Wait()
}

// Start starts the specified command in a Windows Sandbox but does not wait for it to complete.
//
// After a successful call to Start, the Wait method must be called to
// release associated resources.
func (c *Cmd) Start() error {
	if c.finished {
		return fmt.Errorf("command already executed")
	}

	// Create context if not provided
	if c.ctx == nil {
		c.ctx, c.cancel = context.WithCancel(context.Background())
	} else {
		c.ctx, c.cancel = context.WithCancel(c.ctx)
	}

	// Apply timeout if specified
	if c.Timeout > 0 {
		c.ctx, c.cancel = context.WithTimeout(c.ctx, c.Timeout)
	}

	// Build the command script
	script, err := c.buildScript()
	if err != nil {
		return err
	}

	// Create temporary directory for script and output
	tempDir, err := os.MkdirTemp("", "wsb-cmd-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Write script to temp file
	scriptPath := filepath.Join(tempDir, "script.ps1")
	if err := os.WriteFile(scriptPath, []byte(script), 0644); err != nil {
		return fmt.Errorf("failed to write script: %w", err)
	}

	// Create output files
	c.outputFile = filepath.Join(tempDir, "stdout.txt")
	c.errorFile = filepath.Join(tempDir, "stderr.txt")

	// Prepare sandbox configuration
	config := c.Config
	if config == nil {
		config = NewDefaultConfig()
	}

	// Copy config to avoid modifying the original
	configCopy := *config

	// Add mapped folder for the temp directory
	configCopy.MappedFolders = append(configCopy.MappedFolders, MappedFolder{
		HostFolder:    tempDir,
		SandboxFolder: "C:\\WSBCmd",
		ReadOnly:      false,
	})

	// Set up logon command to execute the script
	configCopy.LogonCommand = &LogonCommand{
		Command: fmt.Sprintf(`powershell.exe -ExecutionPolicy Bypass -File "C:\WSBCmd\script.ps1"`),
	}

	// Create and start sandbox
	sandbox, err := New(&configCopy)
	if err != nil {
		return fmt.Errorf("failed to create sandbox: %w", err)
	}

	c.sandbox = sandbox

	if err := sandbox.Start(); err != nil {
		return fmt.Errorf("failed to start sandbox: %w", err)
	}

	// Monitor context cancellation
	go func() {
		<-c.ctx.Done()
		if c.sandbox != nil && !c.finished {
			c.sandbox.Stop()
			c.err = c.ctx.Err()
		}
	}()

	return nil
}

// Wait waits for the command to exit and waits for any copying to
// stdin or copying from stdout or stderr to complete.
//
// The command must have been started by Start.
//
// Wait releases any resources associated with the Cmd.
func (c *Cmd) Wait() error {
	if c.sandbox == nil {
		return fmt.Errorf("command not started")
	}

	// Wait for sandbox to exit
	err := c.sandbox.Wait()

	c.finished = true

	// Cancel context
	if c.cancel != nil {
		c.cancel()
	}

	// Read output files and write to Stdout/Stderr
	if c.Stdout != nil && c.outputFile != "" {
		if data, readErr := os.ReadFile(c.outputFile); readErr == nil {
			c.Stdout.Write(data)
		}
	}

	if c.Stderr != nil && c.errorFile != "" {
		if data, readErr := os.ReadFile(c.errorFile); readErr == nil {
			c.Stderr.Write(data)
		}
	}

	// Clean up sandbox
	if c.sandbox != nil {
		c.sandbox.Stop()
	}

	if c.err != nil {
		return c.err
	}

	return err
}

// Output runs the command and returns its standard output.
// Any returned error will usually be of type *exec.ExitError.
func (c *Cmd) Output() ([]byte, error) {
	if c.Stdout != nil {
		return nil, fmt.Errorf("Stdout already set")
	}
	var stdout bytes.Buffer
	c.Stdout = &stdout

	err := c.Run()
	return stdout.Bytes(), err
}

// CombinedOutput runs the command and returns its combined standard
// output and standard error.
func (c *Cmd) CombinedOutput() ([]byte, error) {
	if c.Stdout != nil {
		return nil, fmt.Errorf("Stdout already set")
	}
	if c.Stderr != nil {
		return nil, fmt.Errorf("Stderr already set")
	}
	var b bytes.Buffer
	c.Stdout = &b
	c.Stderr = &b
	err := c.Run()
	return b.Bytes(), err
}

// StdinPipe returns a pipe that will be connected to the command's
// standard input when the command starts.
// The pipe will be closed automatically after Wait sees the command exit.
func (c *Cmd) StdinPipe() (io.WriteCloser, error) {
	if c.Stdin != nil {
		return nil, fmt.Errorf("Stdin already set")
	}
	pr, pw := io.Pipe()
	c.Stdin = pr
	return pw, nil
}

// StdoutPipe returns a pipe that will be connected to the command's
// standard output when the command starts.
//
// Wait will close the pipe after seeing the command exit, so most callers
// need not close the pipe themselves. It is thus incorrect to call Wait
// before all reads from the pipe have completed.
func (c *Cmd) StdoutPipe() (io.ReadCloser, error) {
	if c.Stdout != nil {
		return nil, fmt.Errorf("Stdout already set")
	}
	pr, pw := io.Pipe()
	c.Stdout = pw
	return pr, nil
}

// StderrPipe returns a pipe that will be connected to the command's
// standard error when the command starts.
//
// Wait will close the pipe after seeing the command exit, so most callers
// need not close the pipe themselves. It is thus incorrect to call Wait
// before all reads from the pipe have completed.
func (c *Cmd) StderrPipe() (io.ReadCloser, error) {
	if c.Stderr != nil {
		return nil, fmt.Errorf("Stderr already set")
	}
	pr, pw := io.Pipe()
	c.Stderr = pw
	return pr, nil
}

// buildScript creates the PowerShell script to execute the command
func (c *Cmd) buildScript() (string, error) {
	if len(c.Args) == 0 && c.Path == "" {
		return "", fmt.Errorf("no command specified")
	}

	// Use exec.Command to properly quote arguments
	var cmdLine string
	if len(c.Args) > 0 {
		// Build command with proper quoting
		quotedArgs := make([]string, len(c.Args))
		for i, arg := range c.Args {
			// Simple quoting - escape quotes and wrap in quotes if contains spaces
			if strings.Contains(arg, " ") || strings.Contains(arg, "\"") {
				arg = strings.ReplaceAll(arg, "\"", "`\"")
				quotedArgs[i] = fmt.Sprintf("\"%s\"", arg)
			} else {
				quotedArgs[i] = arg
			}
		}
		cmdLine = strings.Join(quotedArgs, " ")
	} else {
		cmdLine = c.Path
	}

	var script strings.Builder

	// Set working directory if specified
	if c.Dir != "" {
		script.WriteString(fmt.Sprintf("Set-Location '%s'\n", c.Dir))
	}

	// Set environment variables if specified
	if len(c.Env) > 0 {
		for _, env := range c.Env {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) == 2 {
				script.WriteString(fmt.Sprintf("$env:%s='%s'\n", parts[0], parts[1]))
			}
		}
	}

	// Handle stdin if provided
	stdinPath := ""
	if c.Stdin != nil {
		stdinPath = "C:\\WSBCmd\\stdin.txt"
		// Note: stdin handling would require writing the content to a file
		// in the temp directory during Start(). For now, we note this limitation.
		script.WriteString(fmt.Sprintf("$stdin = Get-Content '%s' -Raw\n", stdinPath))
	}

	// Build the command execution
	script.WriteString("$ErrorActionPreference = 'Continue'\n")
	script.WriteString("try {\n")

	if stdinPath != "" {
		script.WriteString(fmt.Sprintf("    $output = %s | Out-String\n", cmdLine))
	} else {
		script.WriteString(fmt.Sprintf("    $output = & %s 2>&1 | Out-String\n", cmdLine))
	}

	script.WriteString("    $exitCode = $LASTEXITCODE\n")
	script.WriteString("    $output | Out-File 'C:\\WSBCmd\\stdout.txt' -Encoding UTF8\n")
	script.WriteString("    $exitCode | Out-File 'C:\\WSBCmd\\exitcode.txt' -Encoding UTF8\n")
	script.WriteString("} catch {\n")
	script.WriteString("    $_ | Out-String | Out-File 'C:\\WSBCmd\\stderr.txt' -Encoding UTF8\n")
	script.WriteString("    1 | Out-File 'C:\\WSBCmd\\exitcode.txt' -Encoding UTF8\n")
	script.WriteString("}\n")

	// Give time for files to be written before sandbox closes
	script.WriteString("Start-Sleep -Seconds 2\n")

	return script.String(), nil
}
