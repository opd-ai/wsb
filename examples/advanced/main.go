package main

import (
	"fmt"
	"log"
	"os"

	"github.com/opd-ai/wsb/winsandbox"
)

func main() {
	fmt.Println("Windows Sandbox Advanced Example")
	fmt.Println("==================================\n")

	// Check if sandbox is installed
	if err := winsandbox.QuickStart(); err != nil {
		log.Fatal(err)
	}

	// Create advanced configuration
	config := winsandbox.NewDefaultConfig()

	// Enable all features
	config.VGpu = winsandbox.VGpuEnable
	config.Networking = winsandbox.NetworkingEnable
	config.AudioInput = "Enable"
	config.VideoInput = "Enable"
	config.ClipboardRedirection = "Enable"
	config.PrinterRedirection = "Enable"
	config.MemoryInMB = 4096

	// Add mapped folders (example paths - adjust for your system)
	// Note: These folders must exist on the host
	homeDir, _ := os.UserHomeDir()
	if homeDir != "" {
		config.MappedFolders = []winsandbox.MappedFolder{
			{
				HostFolder: homeDir + "\\Downloads",
				ReadOnly:   true,
			},
		}
	}

	// Add startup command
	// This command will run when the sandbox starts
	config.LogonCommand = &winsandbox.LogonCommand{
		Command: "powershell.exe -Command \"Write-Host 'Welcome to Windows Sandbox!' -ForegroundColor Green; Write-Host 'This is an isolated environment.' -ForegroundColor Yellow\"",
	}

	// Save configuration to file
	configFile := "advanced-sandbox.wsb"
	fmt.Printf("Saving configuration to %s...\n", configFile)
	if err := config.WriteToFile(configFile); err != nil {
		log.Fatal(err)
	}
	fmt.Println("✓ Configuration saved")

	// Display configuration
	fmt.Println("\nConfiguration:")
	fmt.Printf("  vGPU: %s\n", config.VGpu)
	fmt.Printf("  Networking: %s\n", config.Networking)
	fmt.Printf("  Memory: %d MB\n", config.MemoryInMB)
	fmt.Printf("  Mapped Folders: %d\n", len(config.MappedFolders))
	if config.LogonCommand != nil {
		fmt.Printf("  Logon Command: %s\n", config.LogonCommand.Command)
	}

	// Launch sandbox
	fmt.Println("\nLaunching Windows Sandbox with advanced configuration...")
	sb, err := winsandbox.New(config)
	if err != nil {
		log.Fatal(err)
	}

	if err := sb.Start(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("✓ Sandbox started successfully!")
	fmt.Println("\nSandbox Features:")
	fmt.Println("  - Hardware acceleration enabled (vGPU)")
	fmt.Println("  - Network connectivity enabled")
	fmt.Println("  - Audio and video input enabled")
	fmt.Println("  - Clipboard sharing enabled")
	fmt.Println("  - Printer redirection enabled")
	if len(config.MappedFolders) > 0 {
		fmt.Println("  - Shared folders mounted")
	}

	// Wait for sandbox to close
	fmt.Println("\nSandbox is running. Close the window to exit.")
	sb.Wait()

	// Cleanup
	sb.Stop()
	fmt.Println("\n✓ Sandbox session ended")

	// Clean up the config file
	os.Remove(configFile)
}
