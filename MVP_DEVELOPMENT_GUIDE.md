# CmdBell MVP Development Guide

## 📋 Project Overview

**CmdBell** is a cross-platform utility that automatically sends notifications when long-running CLI commands complete, supporting both Docker containers and local environments.

**Core Value Proposition:**

- Desktop notifications on Windows host when Docker container (development environment) commands complete
- **Container-internal command detection**: Commands executed directly inside containers (e.g., `sleep 20`) trigger host notifications
- **Hybrid monitoring**: Both Docker events (external) and Shell hooks (internal) supported
- Zero configuration: Ready to use after installation
- Works with all shells and terminals

## 🏗️ Architecture Overview

**CmdBell uses a hybrid monitoring approach:**

### Windows Host (Daemon)
- **Docker Events Monitor**: Tracks `docker run` and `docker exec` commands
- **HTTP Server**: Receives notifications from container Shell hooks (localhost:8080)
- **Native Notifications**: Sends Windows toast notifications

### Container Environment  
- **Shell Integration**: Bash/Zsh/Fish hooks monitor internal command execution
- **HTTP Client**: Sends command completion data to host daemon
- **Auto-Installation**: Automatically installs hooks when container environment detected

## ✅ MVP Completion Status (97% Complete)

### 🎯 Core MVP Features - Completed

- [x] **Docker Container Monitoring** (`docker_monitor.go`)

  - Docker events API integration complete
  - Tracking exec_create, exec_start, exec_die events
  - Container name extraction and command parsing

- [x] **Background Daemon Mode** (`daemon.go`)

  - System service management (start/stop/status/restart)
  - PID file management and process lifecycle
  - Signal handling for clean shutdown

- [x] **Native OS Notifications** (`notification.go`)

  - macOS: Using osascript
  - Linux: notify-send/kdialog/zenity support
  - Windows: PowerShell toast notifications
  - Console fallback for headless environments

- [x] **Shell Integration** (`shell_integration.go`)
  - Automatic integration with bash/zsh/fish
  - preexec/precmd hook implementation
  - Automatic install/uninstall functionality

### ✅ MVP Enhancement Items (Container Internal Commands) - COMPLETED

- [x] **HTTP Communication Server** (`http_server.go`) ✅ **COMPLETED**
  - HTTP server integration in daemon for container notifications
  - POST /notify endpoint for receiving container command data
  - JSON payload handling and validation
  - Health check endpoint (/health)
  - Port 59721 (configurable, avoids common conflicts)

- [ ] **Enhanced Shell Integration** 
  - Modify Shell hooks to use HTTP instead of `--notify` parameter
  - Auto-detect host IP (docker.for.windows.localhost)
  - Fallback to local logging on HTTP failure

- [ ] **Container Environment Detection**
  - Auto-install Shell integration when running inside container
  - Binary deployment and `--install` automation
  - Container-specific configuration handling

### ❌ MVP Missing Items

- [ ] **Unit Tests**: No `*_test.go` files present (to be added later)

## 📂 Codebase Structure

```
src/
├── main.go              # Main entry point, command routing
├── docker_monitor.go    # Docker event monitoring core logic
├── http_server.go       # 🆕 HTTP server for container notifications
├── notification.go      # Cross-platform notification system
├── daemon.go           # Background service management (+ HTTP server integration)
├── shell_integration.go # Shell auto-integration functionality (+ HTTP client)
├── config.go           # Configuration file system
├── go.mod              # Go module definition
└── go.sum              # Dependency checksums

Others/
├── CLAUDE.md           # Project guidelines
├── TESTING.md          # Manual testing guide
└── justfile           # Build automation (currently empty)
```

## 🎯 Development Priorities (Revised Order)

### 🔥 Phase 1: Container Internal Command Support (Current Phase)

> **Goal**: Enable container-internal command detection and notification

- [ ] **HTTP Communication Server** (create new `http_server.go`)
  - HTTP server for receiving container notifications  
  - Integration with existing daemon architecture
  - POST /notify endpoint with JSON payload validation
  - Error handling and logging

- [ ] **Enhanced Shell Integration** (modify `shell_integration.go`)
  - HTTP client functionality in Shell hooks
  - Auto-detect Docker host IP address
  - Fallback mechanisms for network failures
  - Container environment detection

- [ ] **Daemon HTTP Integration** (modify `daemon.go`)
  - Start HTTP server alongside Docker monitoring
  - Unified notification handling for both event sources
  - Configuration for HTTP server port and binding

- [x] **Configuration File System** ✅ **COMPLETED**
  - Support for `.cmdbell.yaml` configuration file
  - Make hardcoded 15-second threshold configurable
  - Notification on/off, per-container settings, etc.
  - Configuration file location: `~/.cmdbell/config.yaml`

### 📦 Phase 2: Deployment and Installation (Completed)

> **Goal**: Install and test on actual PC with CI/CD automation

- [x] **CI/CD Workflow Setup** ✅ **COMPLETED**
  - GitHub Actions for cross-platform builds
  - Automated binary releases for Windows/macOS/Linux
  - Cross-compilation for Windows builds

- [ ] **Package Manager Integration** (Future)
  - Windows: Create Chocolatey package
  - macOS: Create Homebrew formula
  - Linux: Create .deb/.rpm packages

- [ ] **Automatic System Service Registration** (Future)
  - Windows: Register as service
  - macOS: launchd registration
  - Linux: systemd registration

### 🖥️ Phase 3: GUI Application Development (Deferred)

> **Goal**: Enhanced user experience through GUI
> **Note**: Moved to Phase 3 due to Windows GUI build constraints in development environment

- [ ] **GUI Application Development** (new module)
  - Always running via system tray icon
  - Settings GUI: threshold, notification methods, filters, etc.
  - Status monitoring: display currently monitored containers/commands
  - Log viewer: recent notification history
  - Tech stack: Go + Fyne (cross-platform)
  - Build via CI/CD workflow to handle cross-platform GUI compilation

### 🔄 Phase 4: Iterative Development and Improvement

> **Goal**: Gradual improvement through actual usage

- [ ] **Code Refactoring**

  - Separate configuration management module
  - Separate GUI and backend logic
  - Interface-based design for testability

- [ ] **Gradual Test Addition**
  - Core features first: config file loading, Docker event parsing
  - Integration tests: test with actual Docker containers
  - Full tests: cross-platform testing

### 🚀 Phase 5: Production Quality

> **Goal**: Achieve complete commercial product level

- [ ] **CI/CD and Automation**
- [ ] **Performance Optimization**
- [ ] **Advanced Features** (webhooks, metrics, etc.)

## 🎯 Immediate Tasks to Start (Current Sprint)

### 🔥 Priority 1: Container Internal Command Support

```bash
# Files to create/modify
src/http_server.go       # New: HTTP server for container notifications
src/daemon.go           # Modify: Integrate HTTP server with existing daemon
src/shell_integration.go # Modify: Add HTTP client to Shell hooks
src/main.go             # Modify: Add container environment detection
```

**Implementation details:**

- HTTP server listening on localhost:8080
- POST /notify endpoint with JSON payload validation
- Shell hooks send HTTP requests instead of local --notify calls
- Auto-detect Docker host IP (docker.for.windows.localhost)
- Fallback to local logging on network failures

### ✅ Priority 2: CI/CD Workflow Setup - COMPLETED

```bash
# Files created
.github/workflows/
├── release.yml      # Cross-platform build workflow ✅
```

**Implemented features:**

- GitHub Actions for automated builds ✅
- Cross-platform binary compilation ✅
- Automated binary releases for Windows/macOS/Linux ✅

### ✅ Priority 3: Configuration File System - COMPLETED

```bash
# Files created
src/config.go           # Configuration file loading/saving logic ✅
~/.cmdbell/config.yaml  # Default configuration file ✅
```

**Implemented features:**

- YAML-based configuration file ✅
- Notification threshold setting (default: 15 seconds) ✅
- Notification on/off toggle ✅
- Docker monitoring settings ✅
- Notification method selection ✅
- Automatic config file creation ✅

## 💡 Development Philosophy and Approach

### Considering AI-Based Development Characteristics

- **Executable first**: Working product takes priority over tests
- **Iterative improvement**: Find and fix issues through actual usage
- **Gradual testing**: Add test code alongside refactoring
- **User-centered**: Whether it's actually usable is the top priority

### Realistic Development Order

1. **Configuration**: User must have actual control ✅
2. **Deployment/Installation**: Must be testable in real environment (Current Priority)
3. **GUI**: Enhanced user experience via CI/CD automation
4. **Refactoring**: Gradual improvement while understanding code
5. **Testing/Optimization**: Stability assurance comes last

## 🛠️ Build and Test Commands

```bash
# Currently available commands
cd src && go build -o cmdbell .              # Linux/macOS build
cd src && ./cmdbell echo "Hello"             # Short command test
cd src && ./cmdbell sleep 16                 # Long command test
cd src && ./cmdbell --daemon start           # Start daemon (Docker + HTTP)
cd src && ./cmdbell --install                # Install shell integration

# HTTP Server Testing (Completed ✅)
# 1. Start daemon with HTTP server
./cmdbell --daemon start

# 2. Health check
curl -X GET http://localhost:59721/health

# 3. Test notification endpoint
curl -X POST http://localhost:59721/notify \
  -H "Content-Type: application/json" \
  -d '{"command": "sleep 20", "container_name": "test", "duration": "20s", "success": true}'

# 4. From container to Windows host
curl -X POST http://docker.for.windows.localhost:59721/notify \
  -H "Content-Type: application/json" \
  -d '{"command": "npm build", "container_name": "dev_container", "duration": "45s", "success": true}'

# CI/CD and deployment
git push origin release                      # Trigger build workflow
git tag v1.0.0 && git push --tags          # Future: Create release
```

## 📝 Notes for Other AI Agents

1. **Current Status**: MVP core features complete, adding container-internal command support
2. **Current Priority**: HTTP server + Enhanced Shell Integration for container-internal commands
3. **Architecture**: Hybrid monitoring (Docker events + HTTP server) for comprehensive command detection
4. **User Requirements**: Container-internal commands (e.g., `sleep 20` in dev containers) must trigger Windows notifications
5. **Tech Stack**: Go + HTTP server + Shell hooks + GitHub Actions for CI/CD
6. **Testing Environment**: Docker development containers + Windows host notifications

## 🔗 Communication Flow

```
Container Internal Command → Shell Hook → HTTP POST → Windows Daemon → Toast Notification
     (sleep 20)           (preexec/precmd)  (localhost:8080)  (cmdbell.exe)     (Windows)

Docker External Command → Docker Events → Windows Daemon → Toast Notification  
  (docker run alpine sleep 20)    (exec_die)     (cmdbell.exe)    (Windows)
```

**Important**: This hybrid approach ensures both external Docker commands and internal container commands are properly monitored and reported to the Windows host.
