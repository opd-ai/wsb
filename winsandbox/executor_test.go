package winsandbox

import (
	"runtime"
	"testing"
)

func TestNew(t *testing.T) {
	config := NewDefaultConfig()

	_, err := New(config)

	if runtime.GOOS != "windows" {
		if err == nil {
			t.Error("Expected error on non-Windows platform")
		}
		return
	}

	// On Windows, it may fail if sandbox is not installed
	if err != nil {
		t.Logf("New() failed (may be expected if sandbox not installed): %v", err)
	}
}

func TestNewWithInvalidConfig(t *testing.T) {
	config := NewDefaultConfig()
	config.VGpu = "Invalid"

	_, err := New(config)
	if err == nil {
		t.Error("Expected error with invalid configuration")
	}
}

func TestSandboxStart(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping sandbox start test on non-Windows platform")
	}

	config := NewDefaultConfig()
	sb, err := New(config)
	if err != nil {
		t.Skip("Cannot create sandbox:", err)
	}

	// We can't actually start the sandbox in CI, but we can test the error handling
	err = sb.Start()
	if err != nil {
		t.Logf("Start failed (expected in CI): %v", err)
	}
}

func TestSandboxStartWhenAlreadyStarted(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping on non-Windows platform")
	}

	config := NewDefaultConfig()
	sb, err := New(config)
	if err != nil {
		t.Skip("Cannot create sandbox:", err)
	}

	// Mark as started
	sb.started = true

	err = sb.Start()
	if err == nil {
		t.Error("Expected error when starting already running sandbox")
	}
}

func TestSandboxStop(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping on non-Windows platform")
	}

	config := NewDefaultConfig()
	sb, err := New(config)
	if err != nil {
		t.Skip("Cannot create sandbox:", err)
	}

	// Stop should not error even if not started
	err = sb.Stop()
	if err != nil {
		t.Errorf("Stop() returned unexpected error: %v", err)
	}
}

func TestSandboxExecute(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping on non-Windows platform")
	}

	config := NewDefaultConfig()
	sb, err := New(config)
	if err != nil {
		t.Skip("Cannot create sandbox:", err)
	}

	// Execute should fail when sandbox is not running
	_, _, err = sb.Execute("echo test")
	if err == nil {
		t.Error("Expected error when executing command in stopped sandbox")
	}

	// Even if started, Execute is not fully implemented
	sb.started = true
	_, _, err = sb.Execute("echo test")
	if err == nil {
		t.Error("Expected error for unimplemented direct execution")
	}
}

func TestSandboxIsRunning(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping on non-Windows platform")
	}

	config := NewDefaultConfig()
	sb, err := New(config)
	if err != nil {
		t.Skip("Cannot create sandbox:", err)
	}

	if sb.IsRunning() {
		t.Error("Expected sandbox to not be running initially")
	}

	sb.started = true
	if !sb.IsRunning() {
		t.Error("Expected sandbox to be marked as running")
	}
}

func TestSandboxWait(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping on non-Windows platform")
	}

	config := NewDefaultConfig()
	sb, err := New(config)
	if err != nil {
		t.Skip("Cannot create sandbox:", err)
	}

	// Wait should fail when sandbox is not running
	err = sb.Wait()
	if err == nil {
		t.Error("Expected error when waiting on stopped sandbox")
	}
}

func TestSandboxGetConfigPath(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping on non-Windows platform")
	}

	config := NewDefaultConfig()
	sb, err := New(config)
	if err != nil {
		t.Skip("Cannot create sandbox:", err)
	}

	path := sb.GetConfigPath()
	if path != "" {
		t.Error("Expected empty config path before start")
	}
}
