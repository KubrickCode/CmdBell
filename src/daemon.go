package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"
)

type Daemon struct {
	monitor    *DockerMonitor
	pidFile    string
	logFile    string
	ctx        context.Context
	cancel     context.CancelFunc
	isRunning  bool
}

func NewDaemon() *Daemon {
	ctx, cancel := context.WithCancel(context.Background())
	homeDir, _ := os.UserHomeDir()
	
	return &Daemon{
		pidFile: filepath.Join(homeDir, ".cmdbell.pid"),
		logFile: filepath.Join(homeDir, ".cmdbell.log"),
		ctx:     ctx,
		cancel:  cancel,
	}
}

func (d *Daemon) Start() error {
	// Check if already running
	if d.IsRunning() {
		return fmt.Errorf("cmdbell daemon is already running (PID: %d)", d.GetPID())
	}

	// Create Docker monitor
	monitor, err := NewDockerMonitor()
	if err != nil {
		return fmt.Errorf("failed to create Docker monitor: %v", err)
	}
	d.monitor = monitor

	// Write PID file
	if err := d.writePIDFile(); err != nil {
		return fmt.Errorf("failed to write PID file: %v", err)
	}

	// Setup log file
	if err := d.setupLogging(); err != nil {
		return fmt.Errorf("failed to setup logging: %v", err)
	}

	// Start monitoring
	if err := d.monitor.Start(); err != nil {
		d.cleanup()
		return fmt.Errorf("failed to start monitoring: %v", err)
	}

	d.isRunning = true
	log.Println("🚀 CmdBell daemon started successfully")
	
	// Wait for signals
	go d.handleSignals()
	
	return nil
}

func (d *Daemon) Stop() error {
	if !d.IsRunning() {
		return fmt.Errorf("cmdbell daemon is not running")
	}

	pid := d.GetPID()
	if pid == 0 {
		return fmt.Errorf("could not determine daemon PID")
	}

	// Send SIGTERM to the daemon process
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find daemon process: %v", err)
	}

	if err := process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to stop daemon: %v", err)
	}

	// Wait for cleanup
	time.Sleep(1 * time.Second)
	
	// Force cleanup if PID file still exists
	if d.IsRunning() {
		d.cleanup()
	}

	fmt.Println("🛑 CmdBell daemon stopped")
	return nil
}

func (d *Daemon) Status() {
	if d.IsRunning() {
		fmt.Printf("✅ CmdBell daemon is running (PID: %d)\n", d.GetPID())
	} else {
		fmt.Println("❌ CmdBell daemon is not running")
	}
}

func (d *Daemon) IsRunning() bool {
	pid := d.GetPID()
	if pid == 0 {
		return false
	}

	// Check if process actually exists
	process, err := os.FindProcess(pid)
	if err != nil {
		d.cleanup() // Cleanup stale PID file
		return false
	}

	// Send signal 0 to check if process is alive
	err = process.Signal(syscall.Signal(0))
	if err != nil {
		d.cleanup() // Cleanup stale PID file
		return false
	}

	return true
}

func (d *Daemon) GetPID() int {
	data, err := os.ReadFile(d.pidFile)
	if err != nil {
		return 0
	}

	pid, err := strconv.Atoi(string(data))
	if err != nil {
		return 0
	}

	return pid
}

func (d *Daemon) writePIDFile() error {
	pid := os.Getpid()
	return os.WriteFile(d.pidFile, []byte(strconv.Itoa(pid)), 0644)
}

func (d *Daemon) setupLogging() error {
	logFile, err := os.OpenFile(d.logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	log.SetOutput(logFile)
	log.SetFlags(log.LstdFlags)
	return nil
}

func (d *Daemon) handleSignals() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigChan:
		log.Printf("Received signal: %v", sig)
		d.shutdown()
	case <-d.ctx.Done():
		d.shutdown()
	}
}

func (d *Daemon) shutdown() {
	log.Println("🛑 Shutting down CmdBell daemon...")
	
	if d.monitor != nil {
		d.monitor.Stop()
	}
	
	d.cleanup()
	d.cancel()
	d.isRunning = false
	
	log.Println("✅ CmdBell daemon shutdown complete")
	os.Exit(0)
}

func (d *Daemon) cleanup() {
	// Remove PID file
	if err := os.Remove(d.pidFile); err != nil && !os.IsNotExist(err) {
		log.Printf("Failed to remove PID file: %v", err)
	}
}