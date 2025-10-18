# Command Execution Quick Start

This guide demonstrates how to use the Windows Sandbox Command API, which provides an interface similar to Go's `os/exec` package for executing commands inside Windows Sandbox instances.

## Prerequisites

- Windows 10 Pro/Enterprise/Education build 18305 or later
- Windows Sandbox feature installed
- Go 1.16 or later

## Basic Usage

### Simple Command Execution

The simplest way to execute a command is to use `Command()` and `Output()`:

```go
package main

import (
    "fmt"
    "log"
    "github.com/opd-ai/wsb/winsandbox"
)

func main() {
    // Create a command
    cmd := winsandbox.Command("cmd.exe", "/c", "echo", "Hello, Sandbox!")
    
    // Execute and get output
    output, err := cmd.Output()
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println(string(output))
}
```

### Run vs Start/Wait

There are two patterns for executing commands:

#### Pattern 1: Run (Synchronous)

```go
cmd := winsandbox.Command("program.exe", "arg1", "arg2")
err := cmd.Run() // Starts and waits for completion
```

#### Pattern 2: Start/Wait (Asynchronous)

```go
cmd := winsandbox.Command("program.exe", "arg1", "arg2")

// Start the command
if err := cmd.Start(); err != nil {
    log.Fatal(err)
}

// Do other work here...

// Wait for completion
if err := cmd.Wait(); err != nil {
    log.Fatal(err)
}
```

## Advanced Features

### Environment Variables

```go
cmd := winsandbox.Command("cmd.exe", "/c", "echo", "%MYVAR%")
cmd.Env = []string{
    "MYVAR=Hello World",
    "ANOTHER_VAR=value",
}

output, err := cmd.Output()
```

### Working Directory

```go
cmd := winsandbox.Command("cmd.exe", "/c", "dir")
cmd.Dir = "C:\\Windows\\System32"

output, err := cmd.Output()
```

### Timeout

```go
import "time"

cmd := winsandbox.Command("long-running.exe")
cmd.Timeout = 30 * time.Second

err := cmd.Run() // Will timeout after 30 seconds
```

### Context Support

```go
import (
    "context"
    "time"
)

ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

cmd := winsandbox.CommandContext(ctx, "program.exe", "args...")
err := cmd.Run()
```

### Capturing Output

#### Separate Stdout and Stderr

```go
var stdout, stderr bytes.Buffer

cmd := winsandbox.Command("program.exe")
cmd.Stdout = &stdout
cmd.Stderr = &stderr

err := cmd.Run()

fmt.Printf("Stdout: %s\n", stdout.String())
fmt.Printf("Stderr: %s\n", stderr.String())
```

#### Combined Output

```go
output, err := cmd.CombinedOutput() // Both stdout and stderr
```

### Using Pipes

#### Stdin Pipe

```go
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
```

#### Stdout Pipe

```go
import "bufio"

cmd := winsandbox.Command("program.exe")

stdout, err := cmd.StdoutPipe()
if err != nil {
    log.Fatal(err)
}

if err := cmd.Start(); err != nil {
    log.Fatal(err)
}

// Read output line by line
scanner := bufio.NewScanner(stdout)
for scanner.Scan() {
    fmt.Println(scanner.Text())
}

cmd.Wait()
```

### Custom Sandbox Configuration

```go
cmd := winsandbox.Command("program.exe")

// Create custom configuration
config := winsandbox.NewDefaultConfig()
config.Networking = winsandbox.NetworkingEnable
config.VGpu = winsandbox.VGpuEnable
config.MemoryInMB = 4096

// Apply configuration
cmd.Config = config

output, err := cmd.Output()
```

## Common Patterns

### PowerShell Scripts

```go
cmd := winsandbox.Command("powershell.exe", 
    "-ExecutionPolicy", "Bypass",
    "-Command", "Get-Process | Select-Object -First 5")

output, err := cmd.Output()
```

### Batch Files

```go
cmd := winsandbox.Command("cmd.exe", "/c", "my-script.bat")
cmd.Dir = "C:\\Scripts"

err := cmd.Run()
```

### File Operations

```go
// Copy a file
cmd := winsandbox.Command("cmd.exe", "/c", "copy", "source.txt", "dest.txt")
err := cmd.Run()

// List directory
cmd := winsandbox.Command("cmd.exe", "/c", "dir", "/b", "C:\\Windows")
output, err := cmd.Output()
```

## Error Handling

```go
cmd := winsandbox.Command("program.exe")

output, err := cmd.Output()
if err != nil {
    if exitErr, ok := err.(*exec.ExitError); ok {
        // Command ran but failed
        fmt.Printf("Exit code: %d\n", exitErr.ExitCode())
    } else {
        // Command couldn't start
        log.Printf("Failed to start command: %v", err)
    }
    return
}

fmt.Println(string(output))
```

## Tips and Best Practices

1. **Always check for errors**: Command execution can fail for many reasons
2. **Use timeouts**: Long-running commands should have timeouts to prevent hanging
3. **Clean up resources**: Use `defer` to ensure cleanup
4. **Context for cancellation**: Use `CommandContext` when you need cancellation support
5. **Test incrementally**: Test commands locally before running in sandbox
6. **Resource limits**: Be aware that sandbox has limited resources by default

## Comparison with os/exec

The Windows Sandbox Command API is designed to be familiar to users of `os/exec`:

| os/exec | winsandbox | Notes |
|---------|------------|-------|
| `exec.Command()` | `winsandbox.Command()` | Create command |
| `exec.CommandContext()` | `winsandbox.CommandContext()` | With context |
| `cmd.Run()` | `cmd.Run()` | Execute and wait |
| `cmd.Start()` | `cmd.Start()` | Start without waiting |
| `cmd.Wait()` | `cmd.Wait()` | Wait for completion |
| `cmd.Output()` | `cmd.Output()` | Get stdout |
| `cmd.CombinedOutput()` | `cmd.CombinedOutput()` | Get stdout+stderr |
| `cmd.StdinPipe()` | `cmd.StdinPipe()` | Get stdin pipe |
| `cmd.StdoutPipe()` | `cmd.StdoutPipe()` | Get stdout pipe |
| `cmd.StderrPipe()` | `cmd.StderrPipe()` | Get stderr pipe |

## Limitations

- Commands are executed via sandbox startup (LogonCommand)
- Each command starts a new sandbox instance
- Longer startup time compared to direct execution
- Interactive commands may have limitations

## Complete Example

See [examples/cmd-exec/main.go](../examples/cmd-exec/main.go) for a complete working example.
