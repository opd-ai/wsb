package winsandbox

import (
	"runtime"
	"testing"
)

func TestIsInstalled(t *testing.T) {
	// This test can only meaningfully run on Windows
	installed, err := IsInstalled()

	if runtime.GOOS != "windows" {
		if err == nil {
			t.Error("Expected error on non-Windows platform")
		}
		return
	}

	// On Windows, we should get a valid result
	if err != nil {
		t.Logf("IsInstalled error (may be expected if sandbox not installed): %v", err)
	}
	t.Logf("Windows Sandbox installed: %v", installed)
}

func TestRequiresElevation(t *testing.T) {
	// Just test that it doesn't panic
	elevated := RequiresElevation()
	t.Logf("Requires elevation: %v", elevated)

	if runtime.GOOS != "windows" {
		if elevated {
			t.Error("Expected false on non-Windows platform")
		}
	}
}

func TestIsAdministrator(t *testing.T) {
	// Just test that it doesn't panic
	admin := IsAdministrator()
	t.Logf("Is administrator: %v", admin)
}

func TestCheckOSCompatibility(t *testing.T) {
	err := CheckOSCompatibility()

	if runtime.GOOS != "windows" {
		if err == nil {
			t.Error("Expected error on non-Windows platform")
		}
		return
	}

	// On Windows, it may pass or fail depending on the version
	if err != nil {
		t.Logf("OS compatibility check failed (may be expected): %v", err)
	}
}

func TestInstall(t *testing.T) {
	// We can't actually test installation without proper Windows environment
	// and elevation, but we can test error handling
	err := Install()

	if runtime.GOOS != "windows" {
		if err == nil {
			t.Error("Expected error on non-Windows platform")
		}
		return
	}

	// On Windows without elevation or if already installed, we expect specific errors
	t.Logf("Install result (expected to fail without elevation): %v", err)
}

func TestUninstall(t *testing.T) {
	// We can't actually test uninstallation without proper Windows environment
	// and elevation, but we can test error handling
	err := Uninstall()

	if runtime.GOOS != "windows" {
		if err == nil {
			t.Error("Expected error on non-Windows platform")
		}
		return
	}

	// On Windows without elevation, we expect specific errors
	t.Logf("Uninstall result (expected to fail without elevation): %v", err)
}
