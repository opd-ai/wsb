package winsandbox

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestNewDefaultConfig(t *testing.T) {
	config := NewDefaultConfig()

	if config.VGpu != VGpuDefault {
		t.Errorf("Expected VGpu to be Default, got %s", config.VGpu)
	}

	if config.Networking != NetworkingDefault {
		t.Errorf("Expected Networking to be Default, got %s", config.Networking)
	}

	if config.AudioInput != "Default" {
		t.Errorf("Expected AudioInput to be Default, got %s", config.AudioInput)
	}

	if len(config.MappedFolders) != 0 {
		t.Errorf("Expected empty MappedFolders, got %d", len(config.MappedFolders))
	}

	if config.LogonCommand != nil {
		t.Error("Expected nil LogonCommand")
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		wantError bool
		errorMsg  string
	}{
		{
			name:      "valid default config",
			config:    NewDefaultConfig(),
			wantError: false,
		},
		{
			name: "invalid vGPU",
			config: &Config{
				VGpu:                 "Invalid",
				Networking:           NetworkingDefault,
				AudioInput:           "Default",
				VideoInput:           "Default",
				ProtectedClient:      "Default",
				PrinterRedirection:   "Default",
				ClipboardRedirection: "Default",
			},
			wantError: true,
			errorMsg:  "VGpu",
		},
		{
			name: "invalid networking",
			config: &Config{
				VGpu:                 VGpuDefault,
				Networking:           "Invalid",
				AudioInput:           "Default",
				VideoInput:           "Default",
				ProtectedClient:      "Default",
				PrinterRedirection:   "Default",
				ClipboardRedirection: "Default",
			},
			wantError: true,
			errorMsg:  "Networking",
		},
		{
			name: "empty mapped folder host path",
			config: &Config{
				VGpu:                 VGpuDefault,
				Networking:           NetworkingDefault,
				MappedFolders:        []MappedFolder{{HostFolder: ""}},
				AudioInput:           "Default",
				VideoInput:           "Default",
				ProtectedClient:      "Default",
				PrinterRedirection:   "Default",
				ClipboardRedirection: "Default",
			},
			wantError: true,
			errorMsg:  "HostFolder",
		},
		{
			name: "relative mapped folder path",
			config: &Config{
				VGpu:                 VGpuDefault,
				Networking:           NetworkingDefault,
				MappedFolders:        []MappedFolder{{HostFolder: "relative/path"}},
				AudioInput:           "Default",
				VideoInput:           "Default",
				ProtectedClient:      "Default",
				PrinterRedirection:   "Default",
				ClipboardRedirection: "Default",
			},
			wantError: true,
			errorMsg:  "absolute path",
		},
		{
			name: "negative memory",
			config: &Config{
				VGpu:                 VGpuDefault,
				Networking:           NetworkingDefault,
				AudioInput:           "Default",
				VideoInput:           "Default",
				ProtectedClient:      "Default",
				PrinterRedirection:   "Default",
				ClipboardRedirection: "Default",
				MemoryInMB:           -1,
			},
			wantError: true,
			errorMsg:  "MemoryInMB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain '%s', got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			}
		})
	}
}

func TestConfigToWSB(t *testing.T) {
	config := NewDefaultConfig()
	config.VGpu = VGpuEnable
	config.Networking = NetworkingEnable

	data, err := config.ToWSB()
	if err != nil {
		t.Fatalf("ToWSB() failed: %v", err)
	}

	xmlStr := string(data)

	// Check for XML declaration
	if !strings.Contains(xmlStr, "<?xml") {
		t.Error("Expected XML declaration")
	}

	// Check for Configuration root element
	if !strings.Contains(xmlStr, "<Configuration>") {
		t.Error("Expected <Configuration> element")
	}

	// Check for VGpu element
	if !strings.Contains(xmlStr, "<VGpu>Enable</VGpu>") {
		t.Error("Expected <VGpu>Enable</VGpu> element")
	}

	// Check for Networking element
	if !strings.Contains(xmlStr, "<Networking>Enable</Networking>") {
		t.Error("Expected <Networking>Enable</Networking> element")
	}
}

func TestConfigToWSBWithMappedFolders(t *testing.T) {
	config := NewDefaultConfig()
	// Use absolute paths that work on the current OS
	testPath1 := "/tmp/test"
	testPath2 := "/tmp/data"
	if os.PathSeparator == '\\' {
		// Windows
		testPath1 = "C:\\Test"
		testPath2 = "D:\\Data"
	}

	config.MappedFolders = []MappedFolder{
		{HostFolder: testPath1, ReadOnly: true},
		{HostFolder: testPath2, ReadOnly: false, SandboxFolder: "C:\\Users\\WDAGUtilityAccount\\Desktop\\Data"},
	}

	data, err := config.ToWSB()
	if err != nil {
		t.Fatalf("ToWSB() failed: %v", err)
	}

	xmlStr := string(data)

	// Check for MappedFolders element
	if !strings.Contains(xmlStr, "<MappedFolders>") {
		t.Error("Expected <MappedFolders> element")
	}

	// Check for first folder (contains the path or the escaped version)
	if !strings.Contains(xmlStr, "<HostFolder>") {
		t.Error("Expected HostFolder element")
	}

	if !strings.Contains(xmlStr, "<ReadOnly>true</ReadOnly>") {
		t.Error("Expected ReadOnly true element")
	}

	if !strings.Contains(xmlStr, "<ReadOnly>false</ReadOnly>") {
		t.Error("Expected ReadOnly false element")
	}

	if !strings.Contains(xmlStr, "<SandboxFolder>") {
		t.Error("Expected SandboxFolder element")
	}
}

func TestConfigToWSBWithLogonCommand(t *testing.T) {
	config := NewDefaultConfig()
	config.LogonCommand = &LogonCommand{
		Command: "cmd.exe /c echo Hello",
	}

	data, err := config.ToWSB()
	if err != nil {
		t.Fatalf("ToWSB() failed: %v", err)
	}

	xmlStr := string(data)

	// Check for LogonCommand element
	if !strings.Contains(xmlStr, "<LogonCommand>") {
		t.Error("Expected <LogonCommand> element")
	}

	if !strings.Contains(xmlStr, "<Command>cmd.exe /c echo Hello</Command>") {
		t.Error("Expected Command element with correct content")
	}
}

func TestConfigWriteToFile(t *testing.T) {
	config := NewDefaultConfig()
	config.VGpu = VGpuEnable

	// Create temp directory
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.wsb")

	err := config.WriteToFile(filePath)
	if err != nil {
		t.Fatalf("WriteToFile() failed: %v", err)
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("File was not created")
	}

	// Read file and verify content
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	xmlStr := string(data)
	if !strings.Contains(xmlStr, "<VGpu>Enable</VGpu>") {
		t.Error("File content doesn't match expected configuration")
	}
}

func TestConfigWriteToFileWithoutExtension(t *testing.T) {
	config := NewDefaultConfig()

	// Create temp directory
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test")

	err := config.WriteToFile(filePath)
	if err != nil {
		t.Fatalf("WriteToFile() failed: %v", err)
	}

	// Check if file with .wsb extension exists
	expectedPath := filePath + ".wsb"
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Error("File with .wsb extension was not created")
	}
}

func TestErrorTypes(t *testing.T) {
	tests := []struct {
		name  string
		err   error
		check string
	}{
		{
			name:  "ErrNotInstalled",
			err:   ErrNotInstalled{},
			check: "not installed",
		},
		{
			name:  "ErrNotElevated",
			err:   ErrNotElevated{Operation: "test"},
			check: "administrator privileges",
		},
		{
			name:  "ErrIncompatibleOS",
			err:   ErrIncompatibleOS{Version: "Win10", Build: 17000},
			check: "18305",
		},
		{
			name:  "ErrInvalidConfig",
			err:   ErrInvalidConfig{Field: "test", Message: "invalid"},
			check: "test",
		},
		{
			name:  "ErrSandboxFailed",
			err:   ErrSandboxFailed{Operation: "start", Message: "failed"},
			check: "start",
		},
		{
			name:  "ErrPowerShellFailed",
			err:   ErrPowerShellFailed{Command: "test", Output: "error"},
			check: "PowerShell",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errMsg := tt.err.Error()
			if !strings.Contains(errMsg, tt.check) {
				t.Errorf("Expected error message to contain '%s', got: %s", tt.check, errMsg)
			}
		})
	}
}

func TestGetVersion(t *testing.T) {
	version := GetVersion()
	if version == "" {
		t.Error("Expected non-empty version string")
	}

	if !strings.Contains(version, ".") {
		t.Error("Expected version to contain dots (semantic versioning)")
	}
}

func TestVGpuModes(t *testing.T) {
	modes := []VGpuMode{VGpuDefault, VGpuEnable, VGpuDisable}

	for _, mode := range modes {
		config := NewDefaultConfig()
		config.VGpu = mode
		if err := config.Validate(); err != nil {
			t.Errorf("Valid VGpu mode %s failed validation: %v", mode, err)
		}
	}
}

func TestNetworkingModes(t *testing.T) {
	modes := []NetworkingMode{NetworkingDefault, NetworkingEnable, NetworkingDisable}

	for _, mode := range modes {
		config := NewDefaultConfig()
		config.Networking = mode
		if err := config.Validate(); err != nil {
			t.Errorf("Valid Networking mode %s failed validation: %v", mode, err)
		}
	}
}

// Additional tests for better coverage

func TestConfigToWSBMemoryInMB(t *testing.T) {
	config := NewDefaultConfig()
	config.MemoryInMB = 4096

	data, err := config.ToWSB()
	if err != nil {
		t.Fatalf("ToWSB() failed: %v", err)
	}

	xmlStr := string(data)
	if !strings.Contains(xmlStr, "<MemoryInMB>4096</MemoryInMB>") {
		t.Error("Expected MemoryInMB element with value 4096")
	}
}

func TestConfigToWSBAllSettings(t *testing.T) {
	config := NewDefaultConfig()
	config.VGpu = VGpuEnable
	config.Networking = NetworkingEnable
	config.AudioInput = "Enable"
	config.VideoInput = "Enable"
	config.ProtectedClient = "Enable"
	config.PrinterRedirection = "Enable"
	config.ClipboardRedirection = "Enable"

	data, err := config.ToWSB()
	if err != nil {
		t.Fatalf("ToWSB() failed: %v", err)
	}

	xmlStr := string(data)

	// Check all elements are present
	expectedElements := []string{
		"<VGpu>Enable</VGpu>",
		"<Networking>Enable</Networking>",
		"<AudioInput>Enable</AudioInput>",
		"<VideoInput>Enable</VideoInput>",
		"<ProtectedClient>Enable</ProtectedClient>",
		"<PrinterRedirection>Enable</PrinterRedirection>",
		"<ClipboardRedirection>Enable</ClipboardRedirection>",
	}

	for _, elem := range expectedElements {
		if !strings.Contains(xmlStr, elem) {
			t.Errorf("Expected element %s in XML output", elem)
		}
	}
}

func TestConfigValidateAllSettings(t *testing.T) {
	tests := []struct {
		name      string
		modifyFn  func(*Config)
		wantError bool
		errorMsg  string
	}{
		{
			name: "invalid AudioInput",
			modifyFn: func(c *Config) {
				c.AudioInput = "Invalid"
			},
			wantError: true,
			errorMsg:  "AudioInput",
		},
		{
			name: "invalid VideoInput",
			modifyFn: func(c *Config) {
				c.VideoInput = "Invalid"
			},
			wantError: true,
			errorMsg:  "VideoInput",
		},
		{
			name: "invalid ProtectedClient",
			modifyFn: func(c *Config) {
				c.ProtectedClient = "Invalid"
			},
			wantError: true,
			errorMsg:  "ProtectedClient",
		},
		{
			name: "invalid PrinterRedirection",
			modifyFn: func(c *Config) {
				c.PrinterRedirection = "Invalid"
			},
			wantError: true,
			errorMsg:  "PrinterRedirection",
		},
		{
			name: "invalid ClipboardRedirection",
			modifyFn: func(c *Config) {
				c.ClipboardRedirection = "Invalid"
			},
			wantError: true,
			errorMsg:  "ClipboardRedirection",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewDefaultConfig()
			tt.modifyFn(config)
			err := config.Validate()
			if tt.wantError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain '%s', got: %v", tt.errorMsg, err)
				}
			}
		})
	}
}

func TestConfigWriteToFileCreateDirectory(t *testing.T) {
	config := NewDefaultConfig()

	// Create temp directory
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "subdir", "nested", "test.wsb")

	err := config.WriteToFile(filePath)
	if err != nil {
		t.Fatalf("WriteToFile() failed: %v", err)
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("File was not created in nested directory")
	}
}

func TestQuickStart(t *testing.T) {
	// This test will fail on non-Windows or when not elevated,
	// but we can test the error handling
	err := QuickStart()
	// We expect an error on Linux or when not elevated
	if err == nil && runtime.GOOS != "windows" {
		t.Error("Expected error on non-Windows platform")
	}
}

func TestLaunchWithDefaults(t *testing.T) {
	// This test will fail on non-Windows or when sandbox is not installed
	_, err := LaunchWithDefaults()
	// We expect an error on Linux or when sandbox is not installed
	if err == nil && runtime.GOOS != "windows" {
		t.Error("Expected error on non-Windows platform")
	}
}

func TestExample(t *testing.T) {
	// Just run the example function to ensure it doesn't panic
	// It will print output but won't actually do anything on Linux
	Example()
}

func TestConfigEmptyLogonCommand(t *testing.T) {
	config := NewDefaultConfig()
	config.LogonCommand = &LogonCommand{Command: ""}

	data, err := config.ToWSB()
	if err != nil {
		t.Fatalf("ToWSB() failed: %v", err)
	}

	xmlStr := string(data)
	// Empty command should not create LogonCommand element
	if strings.Contains(xmlStr, "<LogonCommand>") {
		t.Error("Expected no LogonCommand element for empty command")
	}
}

func TestMappedFolderWithSandboxFolder(t *testing.T) {
	config := NewDefaultConfig()
	testPath := "/tmp/test"
	if os.PathSeparator == '\\' {
		testPath = "C:\\Test"
	}

	config.MappedFolders = []MappedFolder{
		{
			HostFolder:    testPath,
			SandboxFolder: "C:\\Users\\WDAGUtilityAccount\\Desktop\\Test",
			ReadOnly:      false,
		},
	}

	data, err := config.ToWSB()
	if err != nil {
		t.Fatalf("ToWSB() failed: %v", err)
	}

	xmlStr := string(data)
	if !strings.Contains(xmlStr, "<SandboxFolder>") {
		t.Error("Expected SandboxFolder element")
	}
}
