package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"
)

type DockerEvent struct {
	Type   string           `json:"Type"`
	Action string           `json:"Action"`
	ID     string           `json:"id"`
	Actor  DockerEventActor `json:"Actor"`
	Time   int64            `json:"time"`
}

type DockerEventActor struct {
	ID         string            `json:"ID"`
	Attributes map[string]string `json:"Attributes"`
}

type ContainerExecInfo struct {
	ContainerID   string
	ContainerName string
	Command       string
	StartTime     time.Time
}

type DockerMonitor struct {
	execMap map[string]*ContainerExecInfo
	ctx     context.Context
	cancel  context.CancelFunc
}

func NewDockerMonitor() (*DockerMonitor, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// Check if Docker is available
	cmd := exec.Command("docker", "version")
	if err := cmd.Run(); err != nil {
		cancel()
		return nil, fmt.Errorf("docker is not available: %v", err)
	}

	return &DockerMonitor{
		execMap: make(map[string]*ContainerExecInfo),
		ctx:     ctx,
		cancel:  cancel,
	}, nil
}

func (dm *DockerMonitor) Start() error {
	cmd := exec.CommandContext(dm.ctx, "docker", "events", "--format", "{{json .}}", "--filter", "type=container")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start docker events: %v", err)
	}

	go func() {
		defer cmd.Wait()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			var event DockerEvent
			if err := json.Unmarshal([]byte(line), &event); err != nil {
				log.Printf("Failed to parse Docker event: %v", err)
				continue
			}
			dm.handleEvent(event)
		}
	}()

	fmt.Println("🐳 Docker container monitoring started...")
	return nil
}

func (dm *DockerMonitor) handleEvent(event DockerEvent) {
	if strings.HasPrefix(event.Action, "exec_create:") {
		dm.handleExecCreate(event)
	} else if strings.HasPrefix(event.Action, "exec_start:") {
		dm.handleExecStart(event)
	} else if event.Action == "exec_die" {
		dm.handleExecDie(event)
	}
}

func (dm *DockerMonitor) handleExecCreate(event DockerEvent) {
	execID := event.Actor.Attributes["execID"]
	containerID := event.ID

	// Get container name
	cmd := exec.Command("docker", "inspect", "--format", "{{.Name}}", containerID)
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Failed to get container name for %s: %v", containerID, err)
		return
	}
	containerName := strings.TrimPrefix(strings.TrimSpace(string(output)), "/")

	// Extract command from action (e.g., "exec_create: sleep 17" -> "sleep 17")
	command := "unknown"
	if colonIndex := strings.Index(event.Action, ": "); colonIndex != -1 {
		command = event.Action[colonIndex+2:]
	}

	dm.execMap[execID] = &ContainerExecInfo{
		ContainerID:   containerID,
		ContainerName: containerName,
		Command:       command,
	}

	fmt.Printf("📋 Exec created in container %s (ID: %s)\n", containerName, execID[:12])
}

func (dm *DockerMonitor) handleExecStart(event DockerEvent) {
	execID := event.Actor.Attributes["execID"]
	if info, exists := dm.execMap[execID]; exists {
		info.StartTime = time.Now()
		fmt.Printf("▶️  Command started in container %s\n", info.ContainerName)
	}
}

func (dm *DockerMonitor) handleExecDie(event DockerEvent) {
	execID := event.Actor.Attributes["execID"]
	if info, exists := dm.execMap[execID]; exists {
		duration := time.Since(info.StartTime)
		exitCode := event.Actor.Attributes["exitCode"]
		success := exitCode == "0"

		if globalConfig != nil && duration >= globalConfig.General.MinDurationTime && globalConfig.General.EnableNotify {
			dm.sendContainerNotification(info, duration, success)
		}

		delete(dm.execMap, execID)
		fmt.Printf("🏁 Command completed in container %s (duration: %s, exit: %s)\n",
			info.ContainerName, duration.Round(time.Second), exitCode)
	}
}

func (dm *DockerMonitor) sendContainerNotification(info *ContainerExecInfo, duration time.Duration, success bool) {
	sendContainerNotification(info.Command, info.ContainerName, duration, success)
}

func (dm *DockerMonitor) Stop() {
	dm.cancel()
	fmt.Println("🛑 Docker monitoring stopped")
}
