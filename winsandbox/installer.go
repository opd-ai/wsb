package winsandbox

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

const (
	// MinBuildNumber is the minimum Windows build number that supports Windows Sandbox
	MinBuildNumber = 18305
	// FeatureName is the Windows feature name for Windows Sandbox
	FeatureName = "Containers-DisposableClientVM"
)

// IsInstalled checks if Windows Sandbox feature is installed
// Returns true if installed, false otherwise, and any error encountered
func IsInstalled() (bool, error) {
	if runtime.GOOS != "windows" {
		return false, fmt.Errorf("Windows Sandbox is only supported on Windows")
	}

	// Check if the feature is enabled using DISM
	cmd := exec.Command("powershell.exe", "-ExecutionPolicy", "Bypass", "-NoProfile", "-Command",
		fmt.Sprintf("(Get-WindowsOptionalFeature -Online -FeatureName %s).State -eq 'Enabled'", FeatureName))

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		// If the command fails, try alternative method - check if WindowsSandbox.exe exists
		sandboxPath := os.Getenv("windir")
		if sandboxPath == "" {
			sandboxPath = "C:\\Windows"
		}
		sandboxExe := sandboxPath + "\\System32\\WindowsSandbox.exe"
		if _, err := os.Stat(sandboxExe); err == nil {
			return true, nil
		}
		return false, nil
	}

	output := strings.TrimSpace(stdout.String())
	return output == "True", nil
}

// RequiresElevation checks if the current process requires elevation for installation
// Returns true if not running as administrator
func RequiresElevation() bool {
	if runtime.GOOS != "windows" {
		return false
	}

	// Try to create a file in a protected directory
	testFile := os.Getenv("windir") + "\\Temp\\wsb_elevation_test.tmp"
	f, err := os.Create(testFile)
	if err != nil {
		return true
	}
	f.Close()
	os.Remove(testFile)
	return false
}

// IsAdministrator checks if the current process has administrator privileges
func IsAdministrator() bool {
	return !RequiresElevation()
}

// CheckOSCompatibility verifies that the OS supports Windows Sandbox
func CheckOSCompatibility() error {
	if runtime.GOOS != "windows" {
		return ErrIncompatibleOS{
			Version: runtime.GOOS,
			Build:   0,
		}
	}

	// Get Windows version using PowerShell
	cmd := exec.Command("powershell.exe", "-ExecutionPolicy", "Bypass", "-NoProfile", "-Command",
		"[System.Environment]::OSVersion.Version.Build")

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to detect Windows build: %w", err)
	}

	buildStr := strings.TrimSpace(stdout.String())
	var build int
	_, err = fmt.Sscanf(buildStr, "%d", &build)
	if err != nil {
		return fmt.Errorf("failed to parse Windows build number: %w", err)
	}

	if build < MinBuildNumber {
		return ErrIncompatibleOS{
			Version: fmt.Sprintf("Windows %d", build),
			Build:   build,
		}
	}

	// Check Windows edition
	cmd = exec.Command("powershell.exe", "-ExecutionPolicy", "Bypass", "-NoProfile", "-Command",
		"(Get-WmiObject -Class Win32_OperatingSystem).Caption")

	stdout.Reset()
	cmd.Stdout = &stdout

	err = cmd.Run()
	if err != nil {
		// If we can't detect edition, just warn but don't fail
		return nil
	}

	edition := strings.ToLower(stdout.String())
	if !strings.Contains(edition, "pro") &&
		!strings.Contains(edition, "enterprise") &&
		!strings.Contains(edition, "education") {
		return fmt.Errorf("Windows Sandbox requires Windows 10/11 Pro, Enterprise, or Education edition")
	}

	return nil
}

// Install installs the Windows Sandbox feature
// Requires administrator privileges
func Install() error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("Windows Sandbox is only supported on Windows")
	}

	// Check OS compatibility
	if err := CheckOSCompatibility(); err != nil {
		return err
	}

	// Check if already installed
	installed, err := IsInstalled()
	if err != nil {
		return fmt.Errorf("failed to check installation status: %w", err)
	}
	if installed {
		return nil // Already installed
	}

	// Check for administrator privileges
	if RequiresElevation() {
		return ErrNotElevated{Operation: "Install"}
	}

	// Enable the Windows Sandbox feature
	cmd := exec.Command("powershell.exe", "-ExecutionPolicy", "Bypass", "-NoProfile", "-Command",
		fmt.Sprintf("Enable-WindowsOptionalFeature -FeatureName %s -All -Online -NoRestart", FeatureName))

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Set timeout for installation (5 minutes)
	done := make(chan error, 1)
	go func() {
		done <- cmd.Run()
	}()

	select {
	case err := <-done:
		if err != nil {
			return ErrPowerShellFailed{
				Command: "Enable-WindowsOptionalFeature",
				Output:  stdout.String() + "\n" + stderr.String(),
			}
		}
	case <-time.After(5 * time.Minute):
		cmd.Process.Kill()
		return fmt.Errorf("installation timed out after 5 minutes")
	}

	// Verify installation
	installed, err = IsInstalled()
	if err != nil {
		return fmt.Errorf("failed to verify installation: %w", err)
	}

	if !installed {
		return fmt.Errorf("installation completed but feature is not enabled. A system restart may be required")
	}

	return nil
}

// Uninstall removes the Windows Sandbox feature
// Requires administrator privileges
func Uninstall() error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("Windows Sandbox is only supported on Windows")
	}

	// Check if not installed
	installed, err := IsInstalled()
	if err != nil {
		return fmt.Errorf("failed to check installation status: %w", err)
	}
	if !installed {
		return nil // Already uninstalled
	}

	// Check for administrator privileges
	if RequiresElevation() {
		return ErrNotElevated{Operation: "Uninstall"}
	}

	// Disable the Windows Sandbox feature
	cmd := exec.Command("powershell.exe", "-ExecutionPolicy", "Bypass", "-NoProfile", "-Command",
		fmt.Sprintf("Disable-WindowsOptionalFeature -FeatureName %s -Online -NoRestart", FeatureName))

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		return ErrPowerShellFailed{
			Command: "Disable-WindowsOptionalFeature",
			Output:  stdout.String() + "\n" + stderr.String(),
		}
	}

	return nil
}
