# wsb - Windows Sandbox Go Wrapper

A comprehensive Go package for managing Windows Sandbox, including feature installation, configuration generation, and command execution capabilities.

## Overview

Windows Sandbox is a Windows 10/11 feature that provides isolated desktop environments. This Go wrapper enables programmatic control of Windows Sandbox lifecycle and configuration, making it easy for Go developers to build Windows automation tools.

## Features

- **Feature Installation**: Detect and install Windows Sandbox feature via PowerShell
- **Configuration Management**: Generate valid .wsb (XML) configuration files with full feature support
- **Sandbox Execution**: Launch Windows Sandbox with custom configurations
- **Command Execution**: Execute commands in Windows Sandbox with an API similar to os/exec
- **Error Handling**: Comprehensive error types with actionable messages
- **Type Safety**: Strongly-typed configuration with validation

## Requirements

- Windows 10 Pro/Enterprise/Education build 18305 or later
- Go 1.16 or later
- Administrator privileges for feature installation

## Installation

```bash
go get github.com/opd-ai/wsb
```

## Quick Start

```go
package main

import (
    "log"
    "github.com/opd-ai/wsb/winsandbox"
)

func main() {
    // Check and install if needed
    installed, err := winsandbox.IsInstalled()
    if err != nil {
        log.Fatal(err)
    }
    
    if !installed {
        if err := winsandbox.Install(); err != nil {
            log.Fatal(err)
        }
    }

    // Create configuration
    config := winsandbox.NewDefaultConfig()
    config.Networking = winsandbox.NetworkingEnable
    config.VGpu = winsandbox.VGpuEnable

    // Launch sandbox
    sb, err := winsandbox.New(config)
    if err != nil {
        log.Fatal(err)
    }
    
    if err := sb.Start(); err != nil {
        log.Fatal(err)
    }
    defer sb.Stop()
    
    // Wait for sandbox to close
    sb.Wait()
}
```

## Configuration Options

The package supports all Windows Sandbox configuration parameters:

### Basic Configuration

```go
config := winsandbox.NewDefaultConfig()

// vGPU Configuration
config.VGpu = winsandbox.VGpuEnable    // or VGpuDisable, VGpuDefault

// Networking
config.Networking = winsandbox.NetworkingEnable    // or NetworkingDisable, NetworkingDefault
```

### Mapped Folders

Share folders between host and sandbox:

```go
config.MappedFolders = []winsandbox.MappedFolder{
    {
        HostFolder: "C:\\Data",
        ReadOnly:   true,
    },
    {
        HostFolder:    "C:\\Projects",
        SandboxFolder: "C:\\Users\\WDAGUtilityAccount\\Desktop\\Projects",
        ReadOnly:      false,
    },
}
```

### Logon Command

Execute commands automatically when sandbox starts:

```go
config.LogonCommand = &winsandbox.LogonCommand{
    Command: "powershell.exe -ExecutionPolicy Bypass -File C:\\Scripts\\setup.ps1",
}
```

### Additional Settings

```go
config.AudioInput = "Enable"           // Enable, Disable, or Default
config.VideoInput = "Enable"           // Enable, Disable, or Default
config.ProtectedClient = "Enable"      // Enable, Disable, or Default
config.PrinterRedirection = "Enable"   // Enable, Disable, or Default
config.ClipboardRedirection = "Enable" // Enable, Disable, or Default
config.MemoryInMB = 4096              // Optional memory limit
```

## Configuration File Generation

Generate .wsb configuration files without launching sandbox:

```go
config := winsandbox.NewDefaultConfig()
config.Networking = winsandbox.NetworkingEnable

// Generate XML
xmlData, err := config.ToWSB()
if err != nil {
    log.Fatal(err)
}

// Write to file
err = config.WriteToFile("C:\\configs\\my-sandbox.wsb")
if err != nil {
    log.Fatal(err)
}
```

Example output:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<Configuration>
  <VGpu>Default</VGpu>
  <Networking>Enable</Networking>
  <AudioInput>Default</AudioInput>
  <VideoInput>Default</VideoInput>
  <ProtectedClient>Default</ProtectedClient>
  <PrinterRedirection>Default</PrinterRedirection>
  <ClipboardRedirection>Default</ClipboardRedirection>
</Configuration>
```

## Feature Management API

### Check Installation Status

```go
installed, err := winsandbox.IsInstalled()
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Windows Sandbox installed: %v\n", installed)
```

### Install Windows Sandbox

**Note**: Requires administrator privileges

```go
if winsandbox.RequiresElevation() {
    log.Fatal("Please run as administrator")
}

err := winsandbox.Install()
if err != nil {
    log.Fatal(err)
}
fmt.Println("Windows Sandbox installed successfully")
```

### Check OS Compatibility

```go
err := winsandbox.CheckOSCompatibility()
if err != nil {
    log.Fatal(err)
}
```

## Error Handling

The package provides specific error types for different failure scenarios:

```go
installed, err := winsandbox.IsInstalled()
if err != nil {
    switch e := err.(type) {
    case winsandbox.ErrNotInstalled:
        fmt.Println("Not installed, please install first")
    case winsandbox.ErrNotElevated:
        fmt.Printf("Operation '%s' requires admin privileges\n", e.Operation)
    case winsandbox.ErrIncompatibleOS:
        fmt.Printf("Incompatible OS: %v\n", e)
    default:
        fmt.Printf("Error: %v\n", err)
    }
}
```

## Command Execution API (os/exec-like)

The package provides a command execution API similar to Go's `os/exec` package for running commands inside Windows Sandbox instances.

### Basic Command Execution

```go
// Create a command
cmd := winsandbox.Command("cmd.exe", "/c", "echo", "Hello from sandbox!")

// Execute and get output
output, err := cmd.Output()
if err != nil {
    log.Fatal(err)
}
fmt.Println(string(output))
```

### Command with Environment Variables

```go
cmd := winsandbox.Command("cmd.exe", "/c", "echo", "%MY_VAR%")
cmd.Env = []string{"MY_VAR=Hello World"}

output, err := cmd.Output()
if err != nil {
    log.Fatal(err)
}
```

### Command with Working Directory

```go
cmd := winsandbox.Command("cmd.exe", "/c", "dir")
cmd.Dir = "C:\\Windows\\System32"

output, err := cmd.Output()
if err != nil {
    log.Fatal(err)
}
```

### Command with Context and Timeout

```go
// With timeout
cmd := winsandbox.Command("long-running-program.exe")
cmd.Timeout = 30 * time.Second
err := cmd.Run()

// With context
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

cmd := winsandbox.CommandContext(ctx, "program.exe", "arg1", "arg2")
err := cmd.Run()
```

### Custom Configuration

```go
cmd := winsandbox.Command("powershell.exe", "-Command", "Get-Process")

// Use custom sandbox configuration
config := winsandbox.NewDefaultConfig()
config.Networking = winsandbox.NetworkingEnable
config.VGpu = winsandbox.VGpuEnable
cmd.Config = config

output, err := cmd.Output()
```

### Separate Stdout and Stderr

```go
var stdout, stderr bytes.Buffer

cmd := winsandbox.Command("program.exe")
cmd.Stdout = &stdout
cmd.Stderr = &stderr

err := cmd.Run()
if err != nil {
    log.Printf("Error: %v\nStderr: %s", err, stderr.String())
}
fmt.Printf("Stdout: %s\n", stdout.String())
```

### Using Pipes

```go
// Stdin pipe
cmd := winsandbox.Command("sort.exe")
stdin, err := cmd.StdinPipe()
if err != nil {
    log.Fatal(err)
}

go func() {
    defer stdin.Close()
    io.WriteString(stdin, "zebra\napple\nbanana\n")
}()

output, err := cmd.Output()

// Stdout pipe
cmd := winsandbox.Command("program.exe")
stdout, err := cmd.StdoutPipe()
if err != nil {
    log.Fatal(err)
}

if err := cmd.Start(); err != nil {
    log.Fatal(err)
}

// Read from stdout as it's being produced
scanner := bufio.NewScanner(stdout)
for scanner.Scan() {
    fmt.Println(scanner.Text())
}

cmd.Wait()
```

### Start and Wait Pattern

```go
cmd := winsandbox.Command("program.exe")

// Start the command but don't wait
if err := cmd.Start(); err != nil {
    log.Fatal(err)
}

// Do other work...

// Wait for the command to complete
if err := cmd.Wait(); err != nil {
    log.Fatal(err)
}
```

### Combined Output

```go
// Get both stdout and stderr together
cmd := winsandbox.Command("program.exe")
output, err := cmd.CombinedOutput()
if err != nil {
    log.Fatal(err)
}
fmt.Println(string(output))
```

## Complete Example

```go
package main

import (
    "fmt"
    "log"
    "github.com/opd-ai/wsb/winsandbox"
)

func main() {
    // Check OS compatibility
    if err := winsandbox.CheckOSCompatibility(); err != nil {
        log.Fatal(err)
    }

    // Check if installed
    installed, err := winsandbox.IsInstalled()
    if err != nil {
        log.Fatal(err)
    }

    // Install if needed
    if !installed {
        fmt.Println("Installing Windows Sandbox...")
        if winsandbox.RequiresElevation() {
            log.Fatal("Please run as administrator to install")
        }
        if err := winsandbox.Install(); err != nil {
            log.Fatal(err)
        }
        fmt.Println("Installation complete!")
    }

    // Create custom configuration
    config := winsandbox.NewDefaultConfig()
    config.VGpu = winsandbox.VGpuEnable
    config.Networking = winsandbox.NetworkingEnable
    config.AudioInput = "Enable"
    config.ClipboardRedirection = "Enable"
    
    // Add mapped folders
    config.MappedFolders = []winsandbox.MappedFolder{
        {
            HostFolder: "C:\\SharedData",
            ReadOnly:   true,
        },
    }
    
    // Add startup command
    config.LogonCommand = &winsandbox.LogonCommand{
        Command: "cmd.exe /c echo Welcome to Windows Sandbox!",
    }

    // Validate configuration
    if err := config.Validate(); err != nil {
        log.Fatal(err)
    }

    // Save configuration to file
    if err := config.WriteToFile("my-sandbox.wsb"); err != nil {
        log.Fatal(err)
    }
    fmt.Println("Configuration saved to my-sandbox.wsb")

    // Launch sandbox
    fmt.Println("Launching Windows Sandbox...")
    sb, err := winsandbox.New(config)
    if err != nil {
        log.Fatal(err)
    }

    if err := sb.Start(); err != nil {
        log.Fatal(err)
    }
    fmt.Println("Sandbox started successfully!")

    // Wait for user to close sandbox
    fmt.Println("Sandbox is running. Close the window to exit.")
    if err := sb.Wait(); err != nil {
        log.Printf("Sandbox exited with error: %v", err)
    }

    // Cleanup
    sb.Stop()
    fmt.Println("Sandbox closed.")
}
```

## Security Considerations

- **Privilege Elevation**: Installation requires elevated (administrator) privileges. The package does not automatically attempt re-launch with elevation for security reasons.
- **Mapped Folders**: Data shared via mapped folders is accessible to both host and sandbox. Use read-only mode when possible.
- **LogonCommand**: Commands execute automatically on sandbox startup with sandbox privileges.
- **Configuration Validation**: Always validate configuration before launching sandbox to prevent security misconfigurations.

## API Documentation

Full API documentation is available at [pkg.go.dev](https://pkg.go.dev/github.com/opd-ai/wsb/winsandbox)

### Core Types

- `Config`: Configuration for Windows Sandbox instance
- `Sandbox`: Represents a running Windows Sandbox instance
- `Cmd`: Represents a command to be executed in Windows Sandbox (similar to os/exec.Cmd)
- `MappedFolder`: Folder mapping between host and sandbox
- `LogonCommand`: Command to execute on sandbox startup

### Core Functions

- `IsInstalled() (bool, error)`: Check if Windows Sandbox is installed
- `Install() error`: Install Windows Sandbox feature
- `RequiresElevation() bool`: Check if elevation is required
- `CheckOSCompatibility() error`: Verify OS supports Windows Sandbox
- `NewDefaultConfig() *Config`: Create default configuration
- `New(config *Config) (*Sandbox, error)`: Create new sandbox instance
- `Command(name string, arg ...string) *Cmd`: Create a command to run in sandbox
- `CommandContext(ctx context.Context, name string, arg ...string) *Cmd`: Create a command with context

## Testing

Run tests:

```bash
go test ./winsandbox -v
```

Run tests with coverage:

```bash
go test ./winsandbox -cover
```

**Note**: Some tests require Windows environment and may be skipped on other platforms.

## Contributing

Contributions are welcome! Please ensure:

- Code passes `go fmt`, `go vet`, and all tests
- New features include tests
- Public APIs have godoc comments
- Changes maintain backward compatibility

## License

MIT License - see [LICENSE](LICENSE) file for details

## Acknowledgments

This package wraps the Windows Sandbox feature provided by Microsoft. For more information about Windows Sandbox, visit the [official documentation](https://docs.microsoft.com/en-us/windows/security/threat-protection/windows-sandbox/windows-sandbox-overview).


Donate Monero(The only good cryptocurrency) to support development
==================================================================

 - `monero:43H3Uqnc9rfEsJjUXZYmam45MbtWmREFSANAWY5hijY4aht8cqYaT2BCNhfBhua5XwNdx9Tb6BEdt4tjUHJDwNW5H7mTiwe`

