package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

const MIN_DURATION = 15 * time.Second

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "--monitor":
		startDockerMonitoring()
	case "--daemon":
		handleDaemonCommands()
	case "--install":
		handleShellInstall()
	case "--uninstall":
		handleShellUninstall()
	case "--notify":
		handleNotifyCommand()
	default:
		executeCommand()
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  cmdbell <command> [args...]     - Execute command with notification")
	fmt.Println("  cmdbell --monitor               - Start Docker container monitoring")
	fmt.Println("  cmdbell --daemon start          - Start daemon mode")
	fmt.Println("  cmdbell --daemon stop           - Stop daemon")
	fmt.Println("  cmdbell --daemon status         - Check daemon status")
	fmt.Println("  cmdbell --daemon restart        - Restart daemon")
	fmt.Println("  cmdbell --install               - Install shell integration")
	fmt.Println("  cmdbell --uninstall             - Remove shell integration")
	fmt.Println("  cmdbell --notify <cmd> <dur> <exit> - Internal: send notification")
}

func handleDaemonCommands() {
	if len(os.Args) < 3 {
		fmt.Println("Daemon command required: start, stop, status, restart")
		os.Exit(1)
	}

	daemon := NewDaemon()
	
	switch os.Args[2] {
	case "start":
		if err := daemon.Start(); err != nil {
			fmt.Printf("Failed to start daemon: %v\n", err)
			os.Exit(1)
		}
		
		// Keep running until shutdown
		select {}
		
	case "stop":
		if err := daemon.Stop(); err != nil {
			fmt.Printf("Failed to stop daemon: %v\n", err)
			os.Exit(1)
		}
		
	case "status":
		daemon.Status()
		
	case "restart":
		daemon.Stop() // Ignore error if not running
		time.Sleep(1 * time.Second)
		if err := daemon.Start(); err != nil {
			fmt.Printf("Failed to restart daemon: %v\n", err)
			os.Exit(1)
		}
		
		// Keep running until shutdown
		select {}
		
	default:
		fmt.Println("Invalid daemon command. Use: start, stop, status, restart")
		os.Exit(1)
	}
}

func executeCommand() {
	command := os.Args[1]
	args := os.Args[2:]

	fmt.Printf("Executing: %s %s\n", command, strings.Join(args, " "))
	
	startTime := time.Now()
	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err := cmd.Run()
	duration := time.Since(startTime)

	if duration >= MIN_DURATION {
		sendNotification(command, duration, err == nil)
	}

	if err != nil {
		os.Exit(1)
	}
}

func startDockerMonitoring() {
	monitor, err := NewDockerMonitor()
	if err != nil {
		fmt.Printf("Failed to create Docker monitor: %v\n", err)
		os.Exit(1)
	}

	err = monitor.Start()
	if err != nil {
		fmt.Printf("Failed to start Docker monitoring: %v\n", err)
		os.Exit(1)
	}

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	monitor.Stop()
}

func handleShellInstall() {
	integration, err := NewShellIntegration()
	if err != nil {
		fmt.Printf("Failed to create shell integration: %v\n", err)
		os.Exit(1)
	}

	if err := integration.Install(); err != nil {
		fmt.Printf("Failed to install shell integration: %v\n", err)
		os.Exit(1)
	}
}

func handleShellUninstall() {
	integration, err := NewShellIntegration()
	if err != nil {
		fmt.Printf("Failed to create shell integration: %v\n", err)
		os.Exit(1)
	}

	if err := integration.Uninstall(); err != nil {
		fmt.Printf("Failed to uninstall shell integration: %v\n", err)
		os.Exit(1)
	}
}

func handleNotifyCommand() {
	if len(os.Args) < 5 {
		fmt.Println("Usage: cmdbell --notify <command> <duration_seconds> <exit_code>")
		os.Exit(1)
	}

	command := os.Args[2]
	durationStr := os.Args[3]
	exitCodeStr := os.Args[4]

	duration, err := time.ParseDuration(durationStr + "s")
	if err != nil {
		fmt.Printf("Invalid duration: %v\n", err)
		os.Exit(1)
	}

	success := exitCodeStr == "0"
	sendNotification(command, duration, success)
}

