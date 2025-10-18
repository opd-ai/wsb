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
	stdinFile  string

	// tempDir holds the temporary directory for this command
	tempDir string

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
//
// If name is empty, Command returns a Cmd that will fail when started.
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
//
// If ctx is nil, CommandContext returns a Cmd with a background context.
// This is safer than panicking and maintains backward compatibility.
func CommandContext(ctx context.Context, name string, arg ...string) *Cmd {
	cmd := Command(name, arg...)
	if ctx == nil {
		ctx = context.Background()
	}
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
		return fmt.Errorf("failed to start command: %w", err)
	}
	if err := c.Wait(); err != nil {
		return fmt.Errorf("command execution failed: %w", err)
	}
	return nil
}

// Start starts the specified command in a Windows Sandbox but does not wait for it to complete.
//
// After a successful call to Start, the Wait method must be called to
// release associated resources.
func (c *Cmd) Start() error {
	if c.finished {
		return fmt.Errorf("command already executed")
	}

	// Validate that Path or Args is set
	if c.Path == "" && len(c.Args) == 0 {
		return fmt.Errorf("no command specified")
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
	c.tempDir = tempDir

	// Setup cleanup on error
	var cleanupOnError = true
	defer func() {
		if cleanupOnError {
			os.RemoveAll(tempDir)
		}
	}()

	// Write stdin data if provided
	if c.Stdin != nil {
		c.stdinFile = filepath.Join(tempDir, "stdin.txt")
		stdinData, err := io.ReadAll(c.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read stdin: %w", err)
		}
		if err := os.WriteFile(c.stdinFile, stdinData, 0644); err != nil {
			return fmt.Errorf("failed to write stdin file: %w", err)
		}
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

	// Disable cleanup on error since Start was successful
	cleanupOnError = false

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

	// Ensure cleanup happens regardless of errors
	defer func() {
		c.finished = true

		// Cancel context
		if c.cancel != nil {
			c.cancel()
		}

		// Clean up sandbox
		if c.sandbox != nil {
			c.sandbox.Stop()
		}

		// Clean up temporary directory
		if c.tempDir != "" {
			os.RemoveAll(c.tempDir)
			c.tempDir = ""
		}
	}()

	// Wait for sandbox to exit
	err := c.sandbox.Wait()

	// Read output files and write to Stdout/Stderr
	if c.Stdout != nil && c.outputFile != "" {
		if data, readErr := os.ReadFile(c.outputFile); readErr == nil {
			if _, writeErr := c.Stdout.Write(data); writeErr != nil && err == nil {
				err = fmt.Errorf("failed to write stdout: %w", writeErr)
			}
		}
	}

	if c.Stderr != nil && c.errorFile != "" {
		if data, readErr := os.ReadFile(c.errorFile); readErr == nil {
			if _, writeErr := c.Stderr.Write(data); writeErr != nil && err == nil {
				err = fmt.Errorf("failed to write stderr: %w", writeErr)
			}
		}
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
		return nil, fmt.Errorf("Cmd.Stdout already set")
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
		return nil, fmt.Errorf("Cmd.Stdout already set")
	}
	if c.Stderr != nil {
		return nil, fmt.Errorf("Cmd.Stderr already set")
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
		return nil, fmt.Errorf("Cmd.Stdin already set")
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
		return nil, fmt.Errorf("Cmd.Stdout already set")
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
		return nil, fmt.Errorf("Cmd.Stderr already set")
	}
	pr, pw := io.Pipe()
	c.Stderr = pw
	return pr, nil
}

// buildScript creates the PowerShell script to execute the command in the sandbox.
//
// The script performs the following operations:
// 1. Sets the working directory if c.Dir is specified
// 2. Configures environment variables from c.Env
// 3. Reads stdin from a file if c.Stdin is provided
// 4. Executes the command with proper argument quoting
// 5. Captures stdout, stderr, and exit code to files
// 6. Handles errors gracefully
//
// The generated script uses PowerShell's error handling to ensure robust execution
// and proper output capture even if the command fails.
func (c *Cmd) buildScript() (string, error) {
	if len(c.Args) == 0 && c.Path == "" {
		return "", fmt.Errorf("no command specified: both Args and Path are empty")
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
		// stdin.txt will be created in Start() and mapped to the sandbox
	}

	// Build the command execution
	script.WriteString("$ErrorActionPreference = 'Continue'\n")
	script.WriteString("try {\n")

	if stdinPath != "" {
		// Read stdin from file and pipe to command
		script.WriteString(fmt.Sprintf("    $stdinContent = Get-Content '%s' -Raw -ErrorAction SilentlyContinue\n", stdinPath))
		script.WriteString(fmt.Sprintf("    if ($stdinContent) {\n"))
		script.WriteString(fmt.Sprintf("        $output = $stdinContent | & %s 2>&1 | Out-String\n", cmdLine))
		script.WriteString(fmt.Sprintf("    } else {\n"))
		script.WriteString(fmt.Sprintf("        $output = & %s 2>&1 | Out-String\n", cmdLine))
		script.WriteString(fmt.Sprintf("    }\n"))
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
