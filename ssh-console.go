package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"golang.org/x/crypto/ssh"
)

func main() {
	// Define BMC SSH connection details
	bmcIP := "1.1.1.1" // Replace with your BMC IP
	username := "foo"  // Replace with your username
	password := "bar"  // Replace with your password

	// Create SSH client configuration
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Use proper verification in production
	}

	// Connect to the BMC
	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", bmcIP), config)
	if err != nil {
		log.Fatalf("Failed to connect to BMC: %v", err)
	}
	defer conn.Close()

	// Start a session
	session, err := conn.NewSession()
	if err != nil {
		log.Fatalf("Failed to create SSH session: %v", err)
	}
	defer session.Close()

	// Request a pseudo-terminal for interactive sessions
	err = session.RequestPty("xterm", 80, 40, ssh.TerminalModes{
		ssh.ECHO:          0,     // Disable echo
		ssh.TTY_OP_ISPEED: 14400, // Input speed
		ssh.TTY_OP_OSPEED: 14400, // Output speed
	})
	if err != nil {
		log.Fatalf("Failed to request pseudo-terminal: %v", err)
	}

	// Start the SOL session
	stdin, err := session.StdinPipe()
	if err != nil {
		log.Fatalf("Failed to get stdin pipe: %v", err)
	}
	stdout, err := session.StdoutPipe()
	if err != nil {
		log.Fatalf("Failed to get stdout pipe: %v", err)
	}

	go io.Copy(os.Stdout, stdout) // Stream the SOL output to the terminal

	err = session.Start("console 1") // Replace with your BMC's SOL command
	if err != nil {
		log.Fatalf("Failed to start SOL command: %v", err)
	}

	fmt.Println("Serial-over-LAN session active. Press Ctrl+C to exit.")
	go func() {
		// Allow sending input to the session (optional)
		io.Copy(stdin, os.Stdin)
	}()

	// Wait for the session to end
	err = session.Wait()
	if err != nil {
		log.Fatalf("Error during SOL session: %v", err)
	}
}
