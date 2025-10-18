package winsandbox

import "fmt"

// ErrNotInstalled is returned when Windows Sandbox feature is not installed
type ErrNotInstalled struct{}

func (e ErrNotInstalled) Error() string {
	return "Windows Sandbox feature is not installed. Run Install() with elevated privileges to enable it."
}

// ErrNotElevated is returned when an operation requires administrator privileges
type ErrNotElevated struct {
	Operation string
}

func (e ErrNotElevated) Error() string {
	return fmt.Sprintf("operation '%s' requires administrator privileges. Please run as administrator.", e.Operation)
}

// ErrIncompatibleOS is returned when the OS doesn't support Windows Sandbox
type ErrIncompatibleOS struct {
	Version string
	Build   int
}

func (e ErrIncompatibleOS) Error() string {
	return fmt.Sprintf("Windows Sandbox requires Windows 10 Pro/Enterprise/Education build 18305 or later (detected: %s build %d)", e.Version, e.Build)
}

// ErrInvalidConfig is returned when the sandbox configuration is invalid
type ErrInvalidConfig struct {
	Field   string
	Message string
}

func (e ErrInvalidConfig) Error() string {
	return fmt.Sprintf("invalid configuration for field '%s': %s", e.Field, e.Message)
}

// ErrSandboxFailed is returned when sandbox fails to start or execute
type ErrSandboxFailed struct {
	Operation string
	Message   string
}

func (e ErrSandboxFailed) Error() string {
	return fmt.Sprintf("sandbox %s failed: %s", e.Operation, e.Message)
}

// ErrPowerShellFailed is returned when PowerShell command execution fails
type ErrPowerShellFailed struct {
	Command string
	Output  string
}

func (e ErrPowerShellFailed) Error() string {
	return fmt.Sprintf("PowerShell command failed: %s\nOutput: %s", e.Command, e.Output)
}
