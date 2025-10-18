package winsandbox

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// VGpuMode represents the vGPU configuration mode
type VGpuMode string

const (
	// VGpuDefault uses the default vGPU configuration
	VGpuDefault VGpuMode = "Default"
	// VGpuEnable enables vGPU sharing
	VGpuEnable VGpuMode = "Enable"
	// VGpuDisable disables vGPU sharing
	VGpuDisable VGpuMode = "Disable"
)

// NetworkingMode represents networking configuration
type NetworkingMode string

const (
	// NetworkingEnable enables networking in the sandbox
	NetworkingEnable NetworkingMode = "Enable"
	// NetworkingDisable disables networking in the sandbox
	NetworkingDisable NetworkingMode = "Disable"
	// NetworkingDefault uses the default networking configuration
	NetworkingDefault NetworkingMode = "Default"
)

// MappedFolder represents a folder shared between host and sandbox
type MappedFolder struct {
	HostFolder    string
	SandboxFolder string
	ReadOnly      bool
}

// LogonCommand represents a command to execute on sandbox startup
type LogonCommand struct {
	Command string
}

// Config represents the Windows Sandbox configuration
type Config struct {
	VGpu                 VGpuMode
	Networking           NetworkingMode
	MappedFolders        []MappedFolder
	LogonCommand         *LogonCommand
	AudioInput           string // Enable, Disable, or Default
	VideoInput           string // Enable, Disable, or Default
	ProtectedClient      string // Enable, Disable, or Default
	PrinterRedirection   string // Enable, Disable, or Default
	ClipboardRedirection string // Enable, Disable, or Default
	MemoryInMB           int    // Optional memory limit
}

// xmlConfig represents the XML structure for .wsb files
type xmlConfig struct {
	XMLName              xml.Name          `xml:"Configuration"`
	VGpu                 string            `xml:"VGpu,omitempty"`
	Networking           string            `xml:"Networking,omitempty"`
	MappedFolders        *xmlMappedFolders `xml:"MappedFolders,omitempty"`
	LogonCommand         *xmlLogonCommand  `xml:"LogonCommand,omitempty"`
	AudioInput           string            `xml:"AudioInput,omitempty"`
	VideoInput           string            `xml:"VideoInput,omitempty"`
	ProtectedClient      string            `xml:"ProtectedClient,omitempty"`
	PrinterRedirection   string            `xml:"PrinterRedirection,omitempty"`
	ClipboardRedirection string            `xml:"ClipboardRedirection,omitempty"`
	MemoryInMB           int               `xml:"MemoryInMB,omitempty"`
}

type xmlMappedFolders struct {
	Folders []xmlMappedFolder `xml:"MappedFolder"`
}

type xmlMappedFolder struct {
	HostFolder    string `xml:"HostFolder"`
	SandboxFolder string `xml:"SandboxFolder,omitempty"`
	ReadOnly      string `xml:"ReadOnly"`
}

type xmlLogonCommand struct {
	Command string `xml:"Command"`
}

// NewDefaultConfig creates a new configuration with default values
func NewDefaultConfig() *Config {
	return &Config{
		VGpu:                 VGpuDefault,
		Networking:           NetworkingDefault,
		MappedFolders:        []MappedFolder{},
		LogonCommand:         nil,
		AudioInput:           "Default",
		VideoInput:           "Default",
		ProtectedClient:      "Default",
		PrinterRedirection:   "Default",
		ClipboardRedirection: "Default",
		MemoryInMB:           0, // 0 means use default
	}
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Validate VGpu
	if c.VGpu != VGpuDefault && c.VGpu != VGpuEnable && c.VGpu != VGpuDisable {
		return ErrInvalidConfig{
			Field:   "VGpu",
			Message: "must be 'Default', 'Enable', or 'Disable'",
		}
	}

	// Validate Networking
	if c.Networking != NetworkingDefault && c.Networking != NetworkingEnable && c.Networking != NetworkingDisable {
		return ErrInvalidConfig{
			Field:   "Networking",
			Message: "must be 'Default', 'Enable', or 'Disable'",
		}
	}

	// Validate MappedFolders
	for i, folder := range c.MappedFolders {
		if folder.HostFolder == "" {
			return ErrInvalidConfig{
				Field:   fmt.Sprintf("MappedFolders[%d].HostFolder", i),
				Message: "cannot be empty",
			}
		}
		if !filepath.IsAbs(folder.HostFolder) {
			return ErrInvalidConfig{
				Field:   fmt.Sprintf("MappedFolders[%d].HostFolder", i),
				Message: "must be an absolute path",
			}
		}
	}

	// Validate boolean-style settings
	validSettings := []string{"Default", "Enable", "Disable"}
	settingsToValidate := map[string]string{
		"AudioInput":           c.AudioInput,
		"VideoInput":           c.VideoInput,
		"ProtectedClient":      c.ProtectedClient,
		"PrinterRedirection":   c.PrinterRedirection,
		"ClipboardRedirection": c.ClipboardRedirection,
	}

	for field, value := range settingsToValidate {
		valid := false
		for _, validValue := range validSettings {
			if value == validValue {
				valid = true
				break
			}
		}
		if !valid {
			return ErrInvalidConfig{
				Field:   field,
				Message: "must be 'Default', 'Enable', or 'Disable'",
			}
		}
	}

	// Validate memory
	if c.MemoryInMB < 0 {
		return ErrInvalidConfig{
			Field:   "MemoryInMB",
			Message: "cannot be negative",
		}
	}

	return nil
}

// ToWSB converts the configuration to .wsb XML format
func (c *Config) ToWSB() ([]byte, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}

	xmlCfg := xmlConfig{
		VGpu:                 string(c.VGpu),
		Networking:           string(c.Networking),
		AudioInput:           c.AudioInput,
		VideoInput:           c.VideoInput,
		ProtectedClient:      c.ProtectedClient,
		PrinterRedirection:   c.PrinterRedirection,
		ClipboardRedirection: c.ClipboardRedirection,
		MemoryInMB:           c.MemoryInMB,
	}

	// Convert MappedFolders
	if len(c.MappedFolders) > 0 {
		folders := make([]xmlMappedFolder, len(c.MappedFolders))
		for i, folder := range c.MappedFolders {
			readOnly := "false"
			if folder.ReadOnly {
				readOnly = "true"
			}
			folders[i] = xmlMappedFolder{
				HostFolder:    folder.HostFolder,
				SandboxFolder: folder.SandboxFolder,
				ReadOnly:      readOnly,
			}
		}
		xmlCfg.MappedFolders = &xmlMappedFolders{Folders: folders}
	}

	// Convert LogonCommand
	if c.LogonCommand != nil && c.LogonCommand.Command != "" {
		xmlCfg.LogonCommand = &xmlLogonCommand{
			Command: c.LogonCommand.Command,
		}
	}

	// Marshal to XML
	output, err := xml.MarshalIndent(xmlCfg, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal configuration to XML: %w", err)
	}

	// Add XML header
	xmlHeader := []byte(xml.Header)
	result := append(xmlHeader, output...)
	return result, nil
}

// WriteToFile writes the configuration to a .wsb file
func (c *Config) WriteToFile(path string) error {
	data, err := c.ToWSB()
	if err != nil {
		return err
	}

	// Ensure the file has .wsb extension
	if !strings.HasSuffix(strings.ToLower(path), ".wsb") {
		path = path + ".wsb"
	}

	// Create parent directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write configuration file: %w", err)
	}

	return nil
}
