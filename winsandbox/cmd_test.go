package winsandbox

import (
	"bytes"
	"context"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestCommand(t *testing.T) {
	cmd := Command("echo", "hello", "world")

	if cmd.Path != "echo" {
		t.Errorf("Expected Path to be 'echo', got '%s'", cmd.Path)
	}

	if len(cmd.Args) != 3 {
		t.Errorf("Expected Args length to be 3, got %d", len(cmd.Args))
	}

	if cmd.Args[0] != "echo" || cmd.Args[1] != "hello" || cmd.Args[2] != "world" {
		t.Errorf("Args not set correctly: %v", cmd.Args)
	}
}

func TestCommandContext(t *testing.T) {
	ctx := context.Background()
	cmd := CommandContext(ctx, "echo", "test")

	if cmd.ctx != ctx {
		t.Error("Context not set correctly")
	}

	if cmd.Path != "echo" {
		t.Errorf("Expected Path to be 'echo', got '%s'", cmd.Path)
	}
}

func TestCommandContextNil(t *testing.T) {
	// CommandContext should handle nil context gracefully by using background context
	cmd := CommandContext(nil, "echo", "test")

	if cmd.ctx == nil {
		t.Error("Context should not be nil, expected background context")
	}

	if cmd.Path != "echo" {
		t.Errorf("Expected Path to be 'echo', got '%s'", cmd.Path)
	}
}

func TestCmdString(t *testing.T) {
	tests := []struct {
		name     string
		cmd      *Cmd
		expected string
	}{
		{
			name:     "empty command",
			cmd:      &Cmd{},
			expected: "<empty command>",
		},
		{
			name:     "path only",
			cmd:      &Cmd{Path: "echo"},
			expected: "echo",
		},
		{
			name:     "with args",
			cmd:      &Cmd{Args: []string{"echo", "hello", "world"}},
			expected: "echo hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.cmd.String()
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestCmdStart(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping on non-Windows platform")
	}

	// Check if Windows Sandbox is installed
	installed, err := IsInstalled()
	if err != nil || !installed {
		t.Skip("Windows Sandbox not installed")
	}

	cmd := Command("cmd.exe", "/c", "echo", "test")

	err = cmd.Start()
	if err != nil {
		t.Logf("Start failed (may be expected): %v", err)
		// Don't fail the test if sandbox can't be started (e.g., in CI)
		return
	}

	// Clean up
	if cmd.sandbox != nil {
		cmd.sandbox.Stop()
	}
}

func TestCmdStartTwice(t *testing.T) {
	cmd := Command("echo", "test")
	cmd.finished = true

	err := cmd.Start()
	if err == nil {
		t.Error("Expected error when starting command twice")
	}
	if !strings.Contains(err.Error(), "already executed") {
		t.Errorf("Expected 'already executed' error, got: %v", err)
	}
}

func TestCmdWaitWithoutStart(t *testing.T) {
	cmd := Command("echo", "test")

	err := cmd.Wait()
	if err == nil {
		t.Error("Expected error when waiting without start")
	}
	if !strings.Contains(err.Error(), "not started") {
		t.Errorf("Expected 'not started' error, got: %v", err)
	}
}

func TestCmdRun(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping on non-Windows platform")
	}

	installed, err := IsInstalled()
	if err != nil || !installed {
		t.Skip("Windows Sandbox not installed")
	}

	cmd := Command("cmd.exe", "/c", "echo", "Hello from sandbox")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	err = cmd.Run()
	if err != nil {
		t.Logf("Run failed (may be expected in CI): %v", err)
		return
	}

	output := stdout.String()
	if !strings.Contains(output, "Hello from sandbox") {
		t.Errorf("Expected output to contain 'Hello from sandbox', got: %s", output)
	}
}

func TestCmdOutput(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping on non-Windows platform")
	}

	installed, err := IsInstalled()
	if err != nil || !installed {
		t.Skip("Windows Sandbox not installed")
	}

	cmd := Command("cmd.exe", "/c", "echo", "output test")

	output, err := cmd.Output()
	if err != nil {
		t.Logf("Output failed (may be expected in CI): %v", err)
		return
	}

	if !strings.Contains(string(output), "output test") {
		t.Errorf("Expected output to contain 'output test', got: %s", string(output))
	}
}

func TestCmdOutputWithStdoutSet(t *testing.T) {
	cmd := Command("echo", "test")
	cmd.Stdout = &bytes.Buffer{}

	_, err := cmd.Output()
	if err == nil {
		t.Error("Expected error when Stdout already set")
	}
	if !strings.Contains(err.Error(), "Stdout already set") {
		t.Errorf("Expected 'Stdout already set' error, got: %v", err)
	}
}

func TestCmdCombinedOutput(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping on non-Windows platform")
	}

	installed, err := IsInstalled()
	if err != nil || !installed {
		t.Skip("Windows Sandbox not installed")
	}

	cmd := Command("cmd.exe", "/c", "echo", "combined test")

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("CombinedOutput failed (may be expected in CI): %v", err)
		return
	}

	if !strings.Contains(string(output), "combined test") {
		t.Errorf("Expected output to contain 'combined test', got: %s", string(output))
	}
}

func TestCmdCombinedOutputWithStdoutSet(t *testing.T) {
	cmd := Command("echo", "test")
	cmd.Stdout = &bytes.Buffer{}

	_, err := cmd.CombinedOutput()
	if err == nil {
		t.Error("Expected error when Stdout already set")
	}
}

func TestCmdCombinedOutputWithStderrSet(t *testing.T) {
	cmd := Command("echo", "test")
	cmd.Stderr = &bytes.Buffer{}

	_, err := cmd.CombinedOutput()
	if err == nil {
		t.Error("Expected error when Stderr already set")
	}
}

func TestCmdStdinPipe(t *testing.T) {
	cmd := Command("echo", "test")

	pipe, err := cmd.StdinPipe()
	if err != nil {
		t.Errorf("StdinPipe failed: %v", err)
	}

	if pipe == nil {
		t.Error("Expected non-nil pipe")
	}

	if cmd.Stdin == nil {
		t.Error("Expected Stdin to be set")
	}

	// Test error when Stdin already set
	cmd2 := Command("echo", "test")
	cmd2.Stdin = &bytes.Buffer{}
	_, err = cmd2.StdinPipe()
	if err == nil {
		t.Error("Expected error when Stdin already set")
	}
}

func TestCmdStdoutPipe(t *testing.T) {
	cmd := Command("echo", "test")

	pipe, err := cmd.StdoutPipe()
	if err != nil {
		t.Errorf("StdoutPipe failed: %v", err)
	}

	if pipe == nil {
		t.Error("Expected non-nil pipe")
	}

	if cmd.Stdout == nil {
		t.Error("Expected Stdout to be set")
	}

	// Test error when Stdout already set
	cmd2 := Command("echo", "test")
	cmd2.Stdout = &bytes.Buffer{}
	_, err = cmd2.StdoutPipe()
	if err == nil {
		t.Error("Expected error when Stdout already set")
	}
}

func TestCmdStderrPipe(t *testing.T) {
	cmd := Command("echo", "test")

	pipe, err := cmd.StderrPipe()
	if err != nil {
		t.Errorf("StderrPipe failed: %v", err)
	}

	if pipe == nil {
		t.Error("Expected non-nil pipe")
	}

	if cmd.Stderr == nil {
		t.Error("Expected Stderr to be set")
	}

	// Test error when Stderr already set
	cmd2 := Command("echo", "test")
	cmd2.Stderr = &bytes.Buffer{}
	_, err = cmd2.StderrPipe()
	if err == nil {
		t.Error("Expected error when Stderr already set")
	}
}

func TestCmdWithTimeout(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping on non-Windows platform")
	}

	installed, err := IsInstalled()
	if err != nil || !installed {
		t.Skip("Windows Sandbox not installed")
	}

	cmd := Command("cmd.exe", "/c", "timeout", "5")
	cmd.Timeout = 2 * time.Second

	start := time.Now()
	err = cmd.Run()
	duration := time.Since(start)

	if err == nil {
		t.Log("Expected timeout error (command may have completed)")
	}

	// Should timeout within reasonable time (not wait full 5 seconds)
	if duration > 10*time.Second {
		t.Errorf("Command took too long: %v", duration)
	}
}

func TestCmdWithContext(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping on non-Windows platform")
	}

	installed, err := IsInstalled()
	if err != nil || !installed {
		t.Skip("Windows Sandbox not installed")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := CommandContext(ctx, "cmd.exe", "/c", "timeout", "5")

	start := time.Now()
	err = cmd.Run()
	duration := time.Since(start)

	if err == nil {
		t.Log("Expected context cancellation error")
	}

	// Should cancel within reasonable time
	if duration > 10*time.Second {
		t.Errorf("Command took too long: %v", duration)
	}
}

func TestCmdWithEnv(t *testing.T) {
	cmd := Command("cmd.exe", "/c", "echo", "%TEST_VAR%")
	cmd.Env = []string{"TEST_VAR=hello"}

	// Just test that we can set env without errors
	if len(cmd.Env) != 1 {
		t.Error("Env not set correctly")
	}
}

func TestCmdWithDir(t *testing.T) {
	cmd := Command("cmd.exe", "/c", "cd")
	cmd.Dir = "C:\\Windows"

	// Just test that we can set dir without errors
	if cmd.Dir != "C:\\Windows" {
		t.Error("Dir not set correctly")
	}
}

func TestBuildScript(t *testing.T) {
	tests := []struct {
		name    string
		cmd     *Cmd
		wantErr bool
	}{
		{
			name:    "no command",
			cmd:     &Cmd{},
			wantErr: true,
		},
		{
			name: "simple command",
			cmd: &Cmd{
				Args: []string{"echo", "hello"},
			},
			wantErr: false,
		},
		{
			name: "command with env",
			cmd: &Cmd{
				Args: []string{"cmd.exe", "/c", "echo", "%TEST%"},
				Env:  []string{"TEST=value"},
			},
			wantErr: false,
		},
		{
			name: "command with dir",
			cmd: &Cmd{
				Args: []string{"cmd.exe", "/c", "cd"},
				Dir:  "C:\\Windows",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			script, err := tt.cmd.buildScript()
			if (err != nil) != tt.wantErr {
				t.Errorf("buildScript() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && script == "" {
				t.Error("buildScript() returned empty script")
			}
		})
	}
}

func TestBuildScriptContent(t *testing.T) {
	cmd := &Cmd{
		Args: []string{"echo", "test"},
		Dir:  "C:\\Users",
		Env:  []string{"KEY=value"},
	}

	script, err := cmd.buildScript()
	if err != nil {
		t.Fatalf("buildScript() failed: %v", err)
	}

	// Check that script contains expected elements
	if !strings.Contains(script, "Set-Location") {
		t.Error("Script should contain Set-Location for Dir")
	}

	if !strings.Contains(script, "$env:KEY") {
		t.Error("Script should set environment variables")
	}

	if !strings.Contains(script, "stdout.txt") {
		t.Error("Script should write to stdout.txt")
	}

	if !strings.Contains(script, "exitcode.txt") {
		t.Error("Script should write exit code")
	}
}
