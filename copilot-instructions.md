# Copilot Instructions for wsb (Windows Sandbox Go Wrapper)

## Project Overview
This is a comprehensive Go package for managing Windows Sandbox programmatically. It provides feature installation, configuration generation (XML), and command execution capabilities with an API similar to os/exec. The package focuses on type safety, comprehensive error handling with custom error types, and a clean, idiomatic Go API for Windows Sandbox lifecycle management.

## Code Organization

### Directory Structure
- **Root directory**: Contains go.mod, README.md, LICENSE, and documentation files
- **winsandbox/**: Main package containing all implementation code (flat structure)
- **examples/**: Demonstration programs organized by use case (basic/, advanced/, cmd-exec/)
- **No internal/ or pkg/ directories**: Single package design for simplicity

### Package Naming
- Single package: `winsandbox` - all code lives in one package
- Import path: `github.com/opd-ai/wsb/winsandbox`
- No sub-packages or internal packages

### File Naming
- **Feature-based naming**: Each file contains related functionality
  - `config.go` - Configuration structures and XML marshaling
  - `errors.go` - Custom error types
  - `sandbox.go` - Core Sandbox type and lifecycle management
  - `installer.go` - Windows Sandbox feature installation
  - `cmd.go` - Command execution API (os/exec-like)
  - `executor.go` - Internal execution helpers
- **Test files**: `*_test.go` pattern (e.g., `config_test.go`, `cmd_test.go`)
- **Example tests**: `example_test.go` - contains Example_* functions for documentation

## Coding Standards

### Error Handling

**Primary Pattern**: Use custom error types as values (not pointers) for domain errors

**Custom Error Types** (defined in errors.go):
```go
// Define custom errors as structs
type ErrNotInstalled struct{}

func (e ErrNotInstalled) Error() string {
	return "Windows Sandbox feature is not installed. Run Install() with elevated privileges to enable it."
}

// With context fields
type ErrInvalidConfig struct {
	Field   string
	Message string
}

func (e ErrInvalidConfig) Error() string {
	return fmt.Sprintf("invalid configuration for field '%s': %s", e.Field, e.Message)
}
```

**When to Return Custom Errors**:
- Return custom error types directly (not wrapped) when they represent the primary failure
- Custom errors should be returned as values, not pointers
```go
// Good
return ErrNotInstalled{}
return ErrInvalidConfig{Field: "VGpu", Message: "invalid value"}

// Bad (don't use pointers for custom errors)
return &ErrNotInstalled{}
```

**When to Wrap Errors**:
- Wrap errors from external calls using `fmt.Errorf` with `%w` verb
```go
// Good - wrapping external errors
if err := os.WriteFile(path, data, 0644); err != nil {
	return fmt.Errorf("failed to write configuration file: %w", err)
}

// Good - adding context to external errors
installed, err := IsInstalled()
if err != nil {
	return fmt.Errorf("failed to check installation status: %w", err)
}
```

**Type Assertions for Error Handling**:
```go
// Users can check specific error types
installed, err := winsandbox.IsInstalled()
if err != nil {
	switch e := err.(type) {
	case winsandbox.ErrNotInstalled:
		fmt.Println("Not installed")
	case winsandbox.ErrNotElevated:
		fmt.Printf("Operation '%s' requires admin\n", e.Operation)
	default:
		fmt.Printf("Error: %v\n", err)
	}
}
```

### Naming Conventions

**Variables**:
- camelCase for local variables: `configPath`, `tempDir`, `sandboxExe`
- Descriptive names, avoid abbreviations unless common: `cmd`, `ctx`, `err` are acceptable
- Boolean variables use positive phrasing: `installed`, `started`, `finished`

**Functions and Methods**:
- Exported functions use PascalCase: `IsInstalled()`, `NewDefaultConfig()`, `CheckOSCompatibility()`
- Unexported functions use camelCase: `buildScript()`, `requiresElevation()`
- Constructor pattern: `New()` for main types, `NewDefault*()` for default constructors
- Verb-first naming: `Validate()`, `Start()`, `Stop()`, `Wait()`

**Constants**:
- PascalCase for exported: `VGpuDefault`, `NetworkingEnable`, `MinBuildNumber`
- ALL_CAPS only for truly constant values: `FeatureName = "Containers-DisposableClientVM"`
- Group related constants:
```go
const (
	VGpuDefault VGpuMode = "Default"
	VGpuEnable  VGpuMode = "Enable"
	VGpuDisable VGpuMode = "Disable"
)
```

**Interfaces**:
- This project uses standard library interfaces (io.Reader, io.Writer, context.Context)
- No custom interfaces defined (prefer concrete types for Windows-specific functionality)
- When creating interfaces, use -er suffix: `Reader`, `Writer`, `Validator`

**Types**:
- Use type aliases for semantic clarity:
```go
type VGpuMode string
type NetworkingMode string
```

### Testing

**Test File Naming**: 
- `*_test.go` for unit tests (e.g., `config_test.go`)
- `example_test.go` for Example_* functions that appear in godoc

**Table-Driven Tests**: Use for multiple test cases with similar logic
```go
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
			name:     "with args",
			cmd:      &Cmd{Args: []string{"echo", "hello"}},
			expected: "echo hello",
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
```

**Platform-Specific Testing**: Use runtime.GOOS checks
```go
func TestIsInstalled(t *testing.T) {
	installed, err := IsInstalled()

	if runtime.GOOS != "windows" {
		if err == nil {
			t.Error("Expected error on non-Windows platform")
		}
		return
	}

	// Windows-specific tests
	t.Logf("Windows Sandbox installed: %v", installed)
}
```

**Mocking Approach**: 
- No external mocking libraries used
- Tests rely on actual system calls (integration tests)
- Use t.Logf() for informational output
- Tests gracefully handle expected failures (e.g., without elevation)

**Coverage Expectations**: 
- Focus on API correctness and error handling
- Integration tests that work on Windows
- Tests should not require elevation or installed features
- Example_* functions serve as both documentation and tests

### Concurrency

**Context Usage**:
- Always accept context for long-running or cancellable operations
- Store context in struct for lifecycle management:
```go
type Sandbox struct {
	ctx    context.Context
	cancel context.CancelFunc
	// ...
}

func New(config *Config) (*Sandbox, error) {
	ctx, cancel := context.WithCancel(context.Background())
	return &Sandbox{
		ctx:    ctx,
		cancel: cancel,
	}, nil
}
```

**Goroutines**:
- Use goroutines for monitoring context cancellation:
```go
go func() {
	<-c.ctx.Done()
	if c.sandbox != nil && !c.finished {
		c.sandbox.Stop()
		c.err = c.ctx.Err()
	}
}()
```
- Always ensure goroutines can exit (no leaks)
- Use channels sparingly - prefer context for cancellation

**Timeouts**:
- Support timeouts via context or duration fields:
```go
type Cmd struct {
	Timeout time.Duration
	// ...
}

if c.Timeout > 0 {
	c.ctx, c.cancel = context.WithTimeout(c.ctx, c.Timeout)
}
```

**Synchronization**:
- Use context.Context for cancellation, not sync primitives
- Minimal direct use of sync package
- State flags for lifecycle management: `started`, `finished` (boolean fields)

### Dependencies

**Adding New Dependencies**:
- This project has ZERO external dependencies (uses only standard library)
- Keep it that way - avoid adding dependencies unless absolutely critical
- Standard library packages used:
  - `encoding/xml` for .wsb file generation
  - `os/exec` for PowerShell and Windows Sandbox execution
  - `context` for cancellation
  - `io` for stream handling

**Preferred Standard Library Patterns**:
- Use `os/exec.Command` for external process execution
- Use `encoding/xml` for XML marshaling/unmarshaling
- Use `fmt.Errorf` with `%w` for error wrapping (Go 1.13+)
- Use `os.MkdirAll`, `os.WriteFile`, `os.ReadFile` (not deprecated ioutil)

**Dependency Injection Pattern**: Not used - concrete implementations only

## Architecture Patterns

### Constructor Pattern
Create instances via `New()` functions that validate and initialize:
```go
// Main constructor - validates config and returns ready-to-use instance
func New(config *Config) (*Sandbox, error) {
	if runtime.GOOS != "windows" {
		return nil, fmt.Errorf("Windows Sandbox is only supported on Windows")
	}

	// Validate before creating instance
	if err := config.Validate(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &Sandbox{
		config: config,
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

// Default constructor for common use cases
func NewDefaultConfig() *Config {
	return &Config{
		VGpu:                 VGpuDefault,
		Networking:           NetworkingDefault,
		MappedFolders:        []MappedFolder{},
		AudioInput:           "Default",
		// ... all fields initialized
	}
}
```

### Configuration as Data Pattern
Configuration is a plain struct that can be serialized to XML:
```go
// Public API struct (idiomatic Go)
type Config struct {
	VGpu                 VGpuMode
	Networking           NetworkingMode
	MappedFolders        []MappedFolder
	LogonCommand         *LogonCommand
	MemoryInMB           int
}

// Internal XML struct (matching Windows Sandbox schema)
type xmlConfig struct {
	XMLName    xml.Name `xml:"Configuration"`
	VGpu       string   `xml:"VGpu,omitempty"`
	Networking string   `xml:"Networking,omitempty"`
	// ...
}

// Conversion method
func (c *Config) ToWSB() ([]byte, error) {
	// Convert from public API to XML representation
}
```

### Lifecycle Management Pattern
Types that manage external resources follow Start/Stop/Wait pattern:
```go
type Sandbox struct {
	cmd     *exec.Cmd
	started bool
}

func (s *Sandbox) Start() error {
	if s.started {
		return fmt.Errorf("sandbox is already running")
	}
	// Start sandbox process
	s.started = true
	return nil
}

func (s *Sandbox) Wait() error {
	// Wait for process to complete
	return s.cmd.Wait()
}

func (s *Sandbox) Stop() error {
	// Cleanup and stop
	if s.cancel != nil {
		s.cancel()
	}
	return nil
}
```

### os/exec-like API Pattern
The Cmd type mirrors os/exec.Cmd for familiarity:
```go
// Factory functions
func Command(name string, arg ...string) *Cmd
func CommandContext(ctx context.Context, name string, arg ...string) *Cmd

// Execution patterns
cmd.Run()           // Start and wait
cmd.Start()         // Start async
cmd.Wait()          // Wait for completion
cmd.Output()        // Get stdout
cmd.CombinedOutput() // Get stdout+stderr

// Configuration
cmd.Env = []string{"VAR=value"}
cmd.Dir = "C:\\WorkDir"
cmd.Timeout = 30 * time.Second
```

### Validation Before Execution
Always validate configuration before using it:
```go
func (c *Config) Validate() error {
	// Validate VGpu
	if c.VGpu != VGpuDefault && c.VGpu != VGpuEnable && c.VGpu != VGpuDisable {
		return ErrInvalidConfig{
			Field:   "VGpu",
			Message: "must be 'Default', 'Enable', or 'Disable'",
		}
	}
	
	// Validate paths are absolute
	for i, folder := range c.MappedFolders {
		if !filepath.IsAbs(folder.HostFolder) {
			return ErrInvalidConfig{
				Field:   fmt.Sprintf("MappedFolders[%d].HostFolder", i),
				Message: "must be an absolute path",
			}
		}
	}
	
	return nil
}

// Call Validate before operations
func (s *Sandbox) Start() error {
	if err := s.config.Validate(); err != nil {
		return err
	}
	// ...
}
```

## Common Pitfalls to Avoid

### Don't Use Pointer Receivers for Custom Errors
Custom error types should NOT use pointer receivers - they're returned by value:
```go
// Bad - pointer receiver
func (e *ErrNotInstalled) Error() string {
	return "..."
}

// Good - value receiver
func (e ErrNotInstalled) Error() string {
	return "..."
}
```

### Don't Forget Platform Checks
This is Windows-only code - always check runtime.GOOS:
```go
// Good - check platform first
func IsInstalled() (bool, error) {
	if runtime.GOOS != "windows" {
		return false, fmt.Errorf("Windows Sandbox is only supported on Windows")
	}
	// ... Windows-specific code
}
```

### Don't Forget to Close Resources
Always clean up temporary files and cancel contexts:
```go
// Good - cleanup pattern
func (s *Sandbox) Stop() error {
	if s.cancel != nil {
		s.cancel()
	}
	if s.configPath != "" {
		os.Remove(s.configPath) // Clean up temp file
	}
	return nil
}
```

### Don't Mix XML Field Names with Go Field Names
Keep them separate - use conversion functions:
```go
// Good - separate public API from XML representation
type Config struct {
	VGpu VGpuMode // Go API uses typed enum
}

type xmlConfig struct {
	VGpu string `xml:"VGpu,omitempty"` // XML uses string
}

func (c *Config) ToWSB() ([]byte, error) {
	xmlCfg := xmlConfig{
		VGpu: string(c.VGpu), // Explicit conversion
	}
	// ...
}
```

### Don't Panic on Invalid Input
Return errors instead of panicking (except for nil context):
```go
// Bad
func Command(name string) *Cmd {
	if name == "" {
		panic("empty command name")
	}
}

// Good
func (c *Cmd) Start() error {
	if len(c.Args) == 0 && c.Path == "" {
		return fmt.Errorf("no command specified")
	}
}

// Exception: nil context is programmer error
func CommandContext(ctx context.Context, name string, arg ...string) *Cmd {
	if ctx == nil {
		panic("nil Context") // Following os/exec convention
	}
	// ...
}
```

### Don't Skip Validation
Always call Validate() before using configuration:
```go
// Bad
func New(config *Config) (*Sandbox, error) {
	// Directly use config without validation
	return &Sandbox{config: config}, nil
}

// Good
func New(config *Config) (*Sandbox, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	return &Sandbox{config: config}, nil
}
```

## Documentation Requirements

### Public APIs
Every exported function, type, and method must have a godoc comment:
```go
// IsInstalled checks if Windows Sandbox feature is installed
// Returns true if installed, false otherwise, and any error encountered
func IsInstalled() (bool, error) {
	// ...
}

// Config represents the Windows Sandbox configuration
type Config struct {
	// VGpu specifies the virtual GPU configuration
	VGpu VGpuMode
	// ...
}
```

### Package Documentation
Package-level documentation in sandbox.go includes:
- Package purpose and overview
- Example usage
- Requirements (OS version, privileges)
- Security considerations
```go
// Package winsandbox provides a Go wrapper for Windows Sandbox, enabling programmatic
// control of Windows Sandbox lifecycle and configuration.
//
// Example usage:
//
//	config := winsandbox.NewDefaultConfig()
//	sb, err := winsandbox.New(config)
//	// ...
//
// Requirements:
//   - Windows 10 Pro/Enterprise/Education build 18305 or later
//   - Administrator privileges for feature installation
package winsandbox
```

### Complex Logic
Add inline comments for:
- Non-obvious Windows-specific behavior
- PowerShell command construction
- Temporary file management
```go
// Give time for files to be written before sandbox closes
script.WriteString("Start-Sleep -Seconds 2\n")

// Ensure the file has .wsb extension
if !strings.HasSuffix(strings.ToLower(path), ".wsb") {
	path = path + ".wsb"
}
```

### Example Functions
Provide Example_* functions that appear in godoc:
```go
// Example_command demonstrates basic command execution in Windows Sandbox.
func Example_command() {
	cmd := winsandbox.Command("cmd.exe", "/c", "echo", "Hello")
	output, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(output))
}
```

### README Updates
When adding features, update:
- Quick Start section with new API examples
- Configuration Options with new fields
- API Documentation section with new types/functions

## Before Submitting Code

- [ ] Run `go fmt ./...` - code must be formatted
- [ ] Run `go vet ./...` - must pass with no warnings
- [ ] Run `go test ./...` - all tests must pass
- [ ] Test on Windows if adding Windows-specific code
- [ ] Add/update godoc comments for public APIs
- [ ] Update README.md if adding new features
- [ ] Add Example_* functions for new public APIs
- [ ] Verify no external dependencies added (check go.mod)
- [ ] Check that custom errors are returned by value (not pointer)
- [ ] Ensure platform checks (runtime.GOOS) are in place
- [ ] Validate error handling follows project patterns
- [ ] Update relevant documentation files (COMMAND_API.md if touching Cmd)

## Code Style Guidelines

### Struct Initialization
Use keyed fields for clarity:
```go
// Good
config := &Config{
	VGpu:       VGpuEnable,
	Networking: NetworkingEnable,
	AudioInput: "Default",
}

// Bad - positional fields are unclear
config := &Config{VGpuEnable, NetworkingEnable, []MappedFolder{}, nil, "Default"}
```

### String Building
Use strings.Builder for efficient string concatenation:
```go
var script strings.Builder
script.WriteString("Set-Location 'C:\\'\n")
script.WriteString("$output = & cmd.exe\n")
return script.String()
```

### File Operations
Use modern os package functions (not ioutil):
```go
// Good
data, err := os.ReadFile(path)
err := os.WriteFile(path, data, 0644)
err := os.MkdirAll(dir, 0755)

// Bad - deprecated
data, err := ioutil.ReadFile(path)
err := ioutil.WriteFile(path, data, 0644)
```

### XML Marshaling
Use struct tags and omitempty for optional fields:
```go
type xmlConfig struct {
	XMLName    xml.Name `xml:"Configuration"`
	VGpu       string   `xml:"VGpu,omitempty"`
	MemoryInMB int      `xml:"MemoryInMB,omitempty"`
}

// Add XML header manually
xmlHeader := []byte(xml.Header)
result := append(xmlHeader, output...)
```

### Boolean String Conversion
Explicit string conversion for XML boolean values:
```go
readOnly := "false"
if folder.ReadOnly {
	readOnly = "true"
}
```

## Security Considerations

### Privilege Elevation
- Never automatically attempt privilege elevation
- Check for elevation and return clear error:
```go
if RequiresElevation() {
	return ErrNotElevated{Operation: "Install"}
}
```

### Path Validation
- Always validate paths are absolute
- Use filepath.IsAbs() for validation
- Clean paths with filepath.Clean() when needed

### Command Injection Prevention
- Properly quote arguments when building PowerShell scripts
- Escape special characters:
```go
if strings.Contains(arg, " ") || strings.Contains(arg, "\"") {
	arg = strings.ReplaceAll(arg, "\"", "`\"")
	quotedArgs[i] = fmt.Sprintf("\"%s\"", arg)
}
```

### Temporary File Cleanup
- Always clean up temporary files in defer or Stop()
- Use unique names to prevent conflicts: `sandbox_%d.wsb`

### Configuration Validation
- Validate all configuration before execution
- Check for security implications of MappedFolders (warn about ReadOnly=false)

## Version Information

- **Go Version**: 1.24.9 (specified in go.mod)
- **Target OS**: Windows 10/11 build 18305+
- **Package Version**: Tracked via const Version in sandbox.go
- **API Stability**: Public API should maintain backward compatibility

## Quick Reference

### Import Statement
```go
import "github.com/opd-ai/wsb/winsandbox"
```

### Common Patterns
```go
// Check and install
installed, err := winsandbox.IsInstalled()
if !installed {
	winsandbox.Install() // Requires admin
}

// Create config
config := winsandbox.NewDefaultConfig()
config.Networking = winsandbox.NetworkingEnable

// Launch sandbox
sb, err := winsandbox.New(config)
sb.Start()
defer sb.Stop()
sb.Wait()

// Execute command
cmd := winsandbox.Command("cmd.exe", "/c", "dir")
output, err := cmd.Output()
```

## Testing Philosophy

- Write tests that work on any platform (with graceful skips)
- Use table-driven tests for variations
- Log informational messages with t.Logf()
- Don't require elevated privileges for tests
- Test error paths, not just happy paths
- Example_* functions serve as living documentation

## When in Doubt

- Look at existing code in the same file for patterns
- Follow standard library conventions (especially os/exec)
- Prefer explicit over implicit
- Return errors, don't panic (except programmer errors like nil context)
- Document why, not just what
- Keep the zero-dependency philosophy
