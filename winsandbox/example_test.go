package winsandbox_test

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/opd-ai/wsb/winsandbox"
)

// Example_command demonstrates basic command execution in Windows Sandbox.
func Example_command() {
	// Create a simple command
	cmd := winsandbox.Command("cmd.exe", "/c", "echo", "Hello from sandbox")

	// Execute and get output
	output, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(output))
}

// Example_commandWithEnv demonstrates command execution with environment variables.
func Example_commandWithEnv() {
	cmd := winsandbox.Command("cmd.exe", "/c", "echo", "%MYVAR%")
	cmd.Env = []string{"MYVAR=Hello World"}

	output, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(output))
}

// Example_commandWithTimeout demonstrates command execution with a timeout.
func Example_commandWithTimeout() {
	cmd := winsandbox.Command("cmd.exe", "/c", "timeout", "5")
	cmd.Timeout = 2 * time.Second

	err := cmd.Run()
	if err != nil {
		// Timeout error expected
		fmt.Println("Command timed out as expected")
	}
}

// Example_commandWithContext demonstrates command execution with context.
func Example_commandWithContext() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := winsandbox.CommandContext(ctx, "cmd.exe", "/c", "timeout", "5")

	err := cmd.Run()
	if err != nil {
		// Context cancellation expected
		fmt.Println("Command cancelled by context")
	}
}

// Example_commandStdoutStderr demonstrates capturing separate stdout and stderr.
func Example_commandStdoutStderr() {
	var stdout, stderr bytes.Buffer

	cmd := winsandbox.Command("program.exe")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error: %v\nStderr: %s\n", err, stderr.String())
	}

	fmt.Printf("Stdout: %s\n", stdout.String())
}

// Example_commandDir demonstrates command execution with a working directory.
func Example_commandDir() {
	cmd := winsandbox.Command("cmd.exe", "/c", "cd")
	cmd.Dir = "C:\\Windows"

	output, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(output))
}

// Example_commandCustomConfig demonstrates using a custom sandbox configuration.
func Example_commandCustomConfig() {
	cmd := winsandbox.Command("program.exe", "arg1", "arg2")

	// Create custom sandbox configuration
	config := winsandbox.NewDefaultConfig()
	config.Networking = winsandbox.NetworkingEnable
	config.VGpu = winsandbox.VGpuEnable
	config.MemoryInMB = 4096

	cmd.Config = config

	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

// Example_commandStartWait demonstrates the Start/Wait pattern.
func Example_commandStartWait() {
	cmd := winsandbox.Command("program.exe")

	// Start the command
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	// Do other work here...
	fmt.Println("Command started, doing other work...")

	// Wait for completion
	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Command completed")
}
