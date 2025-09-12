# CmdBell MVP Development Guide

## 📋 Project Overview

**CmdBell** is a cross-platform utility that automatically sends notifications when long-running CLI commands complete, supporting both Docker containers and local environments.

**Core Value Proposition:**
- Desktop notifications on Windows host when Docker container (development environment) commands complete
- Zero configuration: Ready to use after installation
- Works with all shells and terminals

## ✅ MVP Completion Status (95% Complete)

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

### ❌ MVP Missing Items

- [ ] **Unit Tests**: No `*_test.go` files present (to be added later)

## 📂 Codebase Structure

```
src/
├── main.go              # Main entry point, command routing
├── docker_monitor.go    # Docker event monitoring core logic
├── notification.go      # Cross-platform notification system
├── daemon.go           # Background service management
├── shell_integration.go # Shell auto-integration functionality
├── go.mod              # Go module definition
└── go.sum              # Dependency checksums

Others/
├── CLAUDE.md           # Project guidelines
├── TESTING.md          # Manual testing guide
└── justfile           # Build automation (currently empty)
```

## 🎯 Development Priorities (Revised Order)

### 🔥 Phase 1: Core User Features (Current Phase)
> **Goal**: Create a fully usable product

- [ ] **Configuration File System** (create new `config.go`)
  - Support for `.cmdbell.yaml` configuration file
  - Make hardcoded 15-second threshold configurable
  - Notification on/off, per-container settings, etc.
  - Configuration file location: `~/.cmdbell/config.yaml`

- [ ] **GUI Application Development** (new module)
  - Always running via system tray icon
  - Settings GUI: threshold, notification methods, filters, etc.
  - Status monitoring: display currently monitored containers/commands
  - Log viewer: recent notification history
  - Tech stack: Go + Fyne (cross-platform)

### 📦 Phase 2: Actual Deployment and Installation
> **Goal**: Install and test on actual PC

- [ ] **Package Manager Integration**
  - Windows: Create Chocolatey package
  - macOS: Create Homebrew formula
  - Linux: Create .deb/.rpm packages

- [ ] **Automatic System Service Registration**
  - Windows: Register as service
  - macOS: launchd registration
  - Linux: systemd registration

### 🔄 Phase 3: Iterative Development and Improvement
> **Goal**: Gradual improvement through actual usage

- [ ] **Code Refactoring**
  - Separate configuration management module
  - Separate GUI and backend logic
  - Interface-based design for testability

- [ ] **Gradual Test Addition**
  - Core features first: config file loading, Docker event parsing
  - Integration tests: test with actual Docker containers
  - Full tests: cross-platform testing

### 🚀 Phase 4: Production Quality
> **Goal**: Achieve complete commercial product level

- [ ] **CI/CD and Automation**
- [ ] **Performance Optimization**
- [ ] **Advanced Features** (webhooks, metrics, etc.)

## 🎯 Immediate Tasks to Start (1-2 weeks)

### Priority 1: Configuration File System
```bash
# New files to create
src/config.go           # Configuration file loading/saving logic
~/.cmdbell/config.yaml  # Default configuration file
```

**Implementation details:**
- YAML-based configuration file
- Notification threshold setting (default: 15 seconds)
- Per-container filtering
- Notification method selection

### Priority 2: Basic GUI Application
```bash
# New files to create
src/gui/             # GUI package directory
├── main.go         # GUI main
├── tray.go         # System tray
├── settings.go     # Settings window
└── monitor.go      # Status monitoring
```

**Implementation details:**
- Use Fyne framework
- Always running from system tray
- Provide settings GUI
- Real-time status monitoring

### Priority 3: Windows Chocolatey Package
```bash
# New files to create
chocolatey/
├── cmdbell.nuspec    # Package metadata
├── tools/
│   ├── install.ps1   # Installation script
│   └── uninstall.ps1 # Uninstall script
```

## 💡 Development Philosophy and Approach

### Considering AI-Based Development Characteristics
- **Executable first**: Working product takes priority over tests
- **Iterative improvement**: Find and fix issues through actual usage
- **Gradual testing**: Add test code alongside refactoring
- **User-centered**: Whether it's actually usable is the top priority

### Realistic Development Order
1. **GUI + Configuration**: User must have actual control
2. **Deployment/Installation**: Must be testable in real environment
3. **Refactoring**: Gradual improvement while understanding code
4. **Testing/Optimization**: Stability assurance comes last

## 🛠️ Build and Test Commands

```bash
# Currently available commands
cd src && go build -o cmdbell .              # Linux/macOS build
cd src && ./cmdbell echo "Hello"             # Short command test
cd src && ./cmdbell sleep 16                 # Long command test
cd src && ./cmdbell --daemon start           # Start daemon
cd src && ./cmdbell --install                # Install shell integration

# Future commands to be added
cd src && go run gui/main.go                 # Run GUI
chocolatey pack                              # Create Windows package
brew install --build-from-source ./cmdbell.rb  # macOS installation
```

## 📝 Notes for Other AI Agents

1. **Current Status**: MVP is 95% complete, only unit tests missing
2. **Next Tasks**: Configuration file + GUI are top priority
3. **Development Approach**: Executable first, tests later
4. **User Requirements**: Docker container command notifications on Windows is core
5. **Tech Stack**: Go + Fyne for GUI, YAML for config

**Important**: This project is building a tool that the actual user (developer) will use daily, so practicality should be prioritized over theoretical completeness.