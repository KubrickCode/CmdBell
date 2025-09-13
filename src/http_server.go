package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type HTTPServer struct {
	server *http.Server
	port   int
}

type NotificationRequest struct {
	Command       string `json:"command"`
	ContainerName string `json:"container_name"`
	Duration      string `json:"duration"`
	Success       bool   `json:"success"`
	StartTime     string `json:"start_time"`
}

func NewHTTPServer(port int) *HTTPServer {
	return &HTTPServer{
		port: port,
	}
}

func (hs *HTTPServer) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/notify", hs.handleNotification)
	mux.HandleFunc("/health", hs.handleHealth)

	hs.server = &http.Server{
		Addr:    fmt.Sprintf("localhost:%d", hs.port),
		Handler: mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("🌐 Starting HTTP server on localhost:%d", hs.port)
	
	// Start server in goroutine to not block
	go func() {
		if err := hs.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	return nil
}

func (hs *HTTPServer) Stop() error {
	if hs.server == nil {
		return nil
	}

	log.Println("🛑 Stopping HTTP server...")
	return hs.server.Close()
}

func (hs *HTTPServer) handleNotification(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req NotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Invalid JSON payload: %v", err)
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Command == "" {
		http.Error(w, "Missing required field: command", http.StatusBadRequest)
		return
	}

	if req.Duration == "" {
		http.Error(w, "Missing required field: duration", http.StatusBadRequest)
		return
	}

	// Parse duration
	duration, err := time.ParseDuration(req.Duration)
	if err != nil {
		log.Printf("Invalid duration format: %v", err)
		http.Error(w, "Invalid duration format", http.StatusBadRequest)
		return
	}

	// Set default container name if not provided
	containerName := req.ContainerName
	if containerName == "" {
		containerName = "unknown_container"
	}

	log.Printf("📨 Received notification: command='%s', container='%s', duration=%s, success=%t",
		req.Command, containerName, duration, req.Success)

	// Send notification using existing function
	sendContainerNotification(req.Command, containerName, duration, req.Success)

	// Send success response
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"status":  "success",
		"message": "Notification sent",
	}
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

func (hs *HTTPServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"status": "healthy",
		"server": "cmdbell-http",
		"port":   hs.port,
	}
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode health response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}