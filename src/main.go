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
		fmt.Println("Usage:")
		fmt.Println("  cmdbell <command> [args...]     - Execute command with notification")
		fmt.Println("  cmdbell --monitor               - Start Docker container monitoring")
		os.Exit(1)
	}

	if os.Args[1] == "--monitor" {
		startDockerMonitoring()
		return
	}

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

func sendNotification(command string, duration time.Duration, success bool) {
	status := "completed"
	if !success {
		status = "failed"
	}
	
	message := fmt.Sprintf("Command '%s' %s after %s", 
		command, status, duration.Round(time.Second))
	
	fmt.Printf("\n🔔 CmdBell: %s\n", message)
	
	// TODO: Implement native OS notifications
	// - macOS: osascript -e 'display notification...'
	// - Linux: notify-send
	// - Windows: toast notifications
}