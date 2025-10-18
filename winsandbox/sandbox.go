// Package winsandbox provides a Go wrapper for Windows Sandbox, enabling programmatic
// control of Windows Sandbox lifecycle and configuration.
//
// Windows Sandbox is a Windows 10/11 feature that provides isolated desktop environments.
// This package allows Go developers to detect, install, configure, and launch Windows Sandbox
// instances with custom settings.
//
// Example usage:
//
//	// Check and install if needed
//	if installed, _ := winsandbox.IsInstalled(); !installed {
//	    if err := winsandbox.Install(); err != nil {
//	        log.Fatal(err)
//	    }
//	}
//
//	// Create configuration
//	config := winsandbox.NewDefaultConfig()
//	config.Networking = winsandbox.NetworkingEnable
//	config.MappedFolders = []winsandbox.MappedFolder{
//	    {HostFolder: "C:\\Data", ReadOnly: true},
//	}
//
//	// Launch sandbox
//	sb, err := winsandbox.New(config)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	if err := sb.Start(); err != nil {
//	    log.Fatal(err)
//	}
//	defer sb.Stop()
//
// Requirements:
//   - Windows 10 Pro/Enterprise/Education build 18305 or later
//   - Administrator privileges for feature installation
//   - Windows Sandbox feature must be enabled
//
// Security Considerations:
//   - Installation requires elevated (administrator) privileges
//   - Mapped folders share data between host and sandbox
//   - LogonCommand executes automatically on sandbox startup
//   - Always validate configuration before launching sandbox
package winsandbox

import (
	"fmt"
)

// Version is the current version of the winsandbox package
const Version = "1.0.0"

// GetVersion returns the current version of the package
func GetVersion() string {
	return Version
}

// QuickStart is a convenience function that checks if Windows Sandbox is installed,
// and if not, attempts to install it (requires administrator privileges).
// Returns an error if installation fails or OS is incompatible.
func QuickStart() error {
	installed, err := IsInstalled()
	if err != nil {
		return fmt.Errorf("failed to check installation status: %w", err)
	}

	if !installed {
		if RequiresElevation() {
			return ErrNotElevated{Operation: "QuickStart"}
		}
		if err := Install(); err != nil {
			return fmt.Errorf("failed to install Windows Sandbox: %w", err)
		}
	}

	return nil
}

// LaunchWithDefaults creates and starts a Windows Sandbox instance with default configuration
// This is a convenience function for quick testing and development.
func LaunchWithDefaults() (*Sandbox, error) {
	config := NewDefaultConfig()
	sb, err := New(config)
	if err != nil {
		return nil, err
	}

	if err := sb.Start(); err != nil {
		return nil, err
	}

	return sb, nil
}

// Example demonstrates basic usage of the package
func Example() {
	// Check if Windows Sandbox is installed
	installed, err := IsInstalled()
	if err != nil {
		fmt.Printf("Error checking installation: %v\n", err)
		return
	}

	if !installed {
		fmt.Println("Windows Sandbox is not installed")
		fmt.Println("To install, run with administrator privileges:")
		fmt.Println("  if err := winsandbox.Install(); err != nil { ... }")
		return
	}

	fmt.Println("Windows Sandbox is installed")

	// Create a custom configuration
	config := NewDefaultConfig()
	config.Networking = NetworkingEnable
	config.VGpu = VGpuEnable

	// Add a mapped folder (example)
	// config.MappedFolders = []MappedFolder{
	//     {HostFolder: "C:\\MyFolder", ReadOnly: true},
	// }

	// Create and start sandbox
	sb, err := New(config)
	if err != nil {
		fmt.Printf("Error creating sandbox: %v\n", err)
		return
	}

	fmt.Println("Starting Windows Sandbox...")
	if err := sb.Start(); err != nil {
		fmt.Printf("Error starting sandbox: %v\n", err)
		return
	}

	fmt.Println("Windows Sandbox started successfully")
	fmt.Println("Close the sandbox window to exit")

	// Wait for sandbox to exit
	sb.Wait()
	fmt.Println("Windows Sandbox closed")
}
