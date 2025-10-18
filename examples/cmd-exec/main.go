package main

import (
	"fmt"
	"log"

	"github.com/opd-ai/wsb/winsandbox"
)

func main() {
	fmt.Println("Windows Sandbox Command Execution Example")
	fmt.Println("==========================================")
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
		fmt.Println("\nPlease install Windows Sandbox first:")
		fmt.Println("  - Run as administrator")
		fmt.Println("  - Call winsandbox.Install()")
		return
	}

	fmt.Println("✓ Windows Sandbox is installed")

	// Example 1: Simple command execution using Command
	fmt.Println("\n=== Example 1: Simple Command ===")
	cmd := winsandbox.Command("cmd.exe", "/c", "echo", "Hello from Windows Sandbox!")

	output, err := cmd.Output()
	if err != nil {
		log.Printf("Command failed: %v", err)
	} else {
		fmt.Printf("Output: %s\n", string(output))
	}

	// Example 2: Command with environment variables
	fmt.Println("\n=== Example 2: Command with Environment Variables ===")
	cmd2 := winsandbox.Command("cmd.exe", "/c", "echo", "%CUSTOM_VAR%")
	cmd2.Env = []string{"CUSTOM_VAR=Hello from custom environment!"}

	output2, err := cmd2.Output()
	if err != nil {
		log.Printf("Command failed: %v", err)
	} else {
		fmt.Printf("Output: %s\n", string(output2))
	}

	// Example 3: Command with custom working directory
	fmt.Println("\n=== Example 3: Command with Working Directory ===")
	cmd3 := winsandbox.Command("cmd.exe", "/c", "cd")
	cmd3.Dir = "C:\\Windows\\System32"

	output3, err := cmd3.Output()
	if err != nil {
		log.Printf("Command failed: %v", err)
	} else {
		fmt.Printf("Current Directory: %s\n", string(output3))
	}

	// Example 4: PowerShell command
	fmt.Println("\n=== Example 4: PowerShell Command ===")
	cmd4 := winsandbox.Command("powershell.exe", "-Command", "Get-ComputerInfo | Select-Object CsName, WindowsVersion")

	output4, err := cmd4.Output()
	if err != nil {
		log.Printf("Command failed: %v", err)
	} else {
		fmt.Printf("Computer Info:\n%s\n", string(output4))
	}

	// Example 5: List files in a directory
	fmt.Println("\n=== Example 5: List Files ===")
	cmd5 := winsandbox.Command("cmd.exe", "/c", "dir", "C:\\Windows")

	output5, err := cmd5.Output()
	if err != nil {
		log.Printf("Command failed: %v", err)
	} else {
		fmt.Printf("Directory listing:\n%s\n", string(output5))
	}

	fmt.Println("\n✓ All examples completed!")
}
