package main

import (
	"fmt"
	"log"

	"github.com/opd-ai/wsb/winsandbox"
)

func main() {
	fmt.Println("Windows Sandbox Basic Example")
	fmt.Println("==============================")
	fmt.Println()

	// Check OS compatibility
	fmt.Println("Checking OS compatibility...")
	if err := winsandbox.CheckOSCompatibility(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("✓ OS is compatible")

	// Check if installed
	fmt.Println("\nChecking if Windows Sandbox is installed...")
	installed, err := winsandbox.IsInstalled()
	if err != nil {
		log.Fatal(err)
	}

	if !installed {
		fmt.Println("✗ Windows Sandbox is not installed")
		if winsandbox.RequiresElevation() {
			fmt.Println("\nTo install, please run this program as administrator.")
			fmt.Println("Then the installation will proceed automatically.")
			return
		}

		fmt.Println("\nInstalling Windows Sandbox...")
		if err := winsandbox.Install(); err != nil {
			log.Fatal(err)
		}
		fmt.Println("✓ Installation complete! You may need to restart your computer.")
		return
	}

	fmt.Println("✓ Windows Sandbox is installed")

	// Create a basic configuration
	fmt.Println("\nCreating sandbox configuration...")
	config := winsandbox.NewDefaultConfig()
	config.VGpu = winsandbox.VGpuEnable
	config.Networking = winsandbox.NetworkingEnable
	config.ClipboardRedirection = "Enable"

	// Validate configuration
	if err := config.Validate(); err != nil {
		log.Fatalf("Configuration validation failed: %v", err)
	}
	fmt.Println("✓ Configuration validated")

	// Launch sandbox
	fmt.Println("\nLaunching Windows Sandbox...")
	sb, err := winsandbox.New(config)
	if err != nil {
		log.Fatal(err)
	}

	if err := sb.Start(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("✓ Sandbox started successfully!")

	// Wait for sandbox to close
	fmt.Println("\nSandbox is running. Close the window to exit.")
	if err := sb.Wait(); err != nil {
		log.Printf("Sandbox exited with error: %v", err)
	}

	// Cleanup
	sb.Stop()
	fmt.Println("\n✓ Sandbox closed successfully")
}
