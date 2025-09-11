# CmdBell Testing Guide

## Overview
CmdBell is a cross-platform command completion notifier that monitors Docker containers and local commands, sending native OS notifications when long-running tasks complete.

## Prerequisites
- Docker installed and running
- Go 1.19+ (for building from source)

## Building

### Linux/macOS
```bash
cd src
go build -o cmdbell .
```

### Windows
```bash
cd src
GOOS=windows GOARCH=amd64 go build -o cmdbell.exe .
```

## Test Scenarios

### 1. Basic Command Monitoring

**Short commands (no notification):**
```bash
./cmdbell echo "Hello World"
./cmdbell ls -la
```

**Long commands (15+ seconds, triggers notification):**
```bash
./cmdbell sleep 16
./cmdbell sleep 20
```

**Failed long commands (failure notification):**
```bash
./cmdbell sh -c "sleep 16; exit 1"
```

### 2. Daemon Mode Testing

**Start daemon:**
```bash
./cmdbell --daemon start &
```

**Check status:**
```bash
./cmdbell --daemon status
# Expected: ✅ CmdBell daemon is running (PID: XXXX)
```

**Stop daemon:**
```bash
./cmdbell --daemon stop
```

**Restart daemon:**
```bash
./cmdbell --daemon restart &
```

### 3. Docker Container Monitoring

**List running containers:**
```bash
docker ps
```

**Test with existing containers:**
```bash
# Replace container_name with actual container from docker ps
docker exec <container_name> sleep 18
docker exec <container_name> sh -c "sleep 16; exit 1"  # failure case
```

**Create test container:**
```bash
docker run --rm -d --name test-container ubuntu:20.04 sleep 300
docker exec test-container sleep 20
docker stop test-container
```

### 4. Cross-Platform Testing

#### Linux
```bash
# Install notification dependencies if needed
sudo apt-get install libnotify-bin

# Test with GUI environment
export DISPLAY=:0
./cmdbell sleep 16

# Headless environment (console output only)
unset DISPLAY
./cmdbell sleep 16
```

#### macOS
```bash
./cmdbell sleep 16  # Native macOS notification
```

#### Windows
```powershell
# PowerShell/CMD
.\cmdbell.exe sleep 16  # Windows toast notification
```

### 5. Windows Host + Container Testing

This is the primary use case: monitoring Docker containers from Windows host.

**Setup:**
1. Ensure Docker Desktop is running on Windows
2. Copy `cmdbell.exe` to Windows machine
3. Start daemon on Windows host

**Windows PowerShell:**
```powershell
# Start monitoring daemon
.\cmdbell.exe --daemon start

# Check status in another terminal
.\cmdbell.exe --daemon status
```

**Container (WSL/Dev Container/Remote):**
```bash
# Execute long commands in any container
docker exec <container_id> sleep 20
docker exec <dev_container_name> npm run build
docker exec <dev_container_name> go test ./...
```

**Expected Result:**
- Windows toast notifications appear when container commands complete
- Notifications show container name, command, duration, and success/failure status

### 6. Development Workflow Testing

**VS Code Dev Container scenario:**
1. Start cmdbell daemon on Windows host
2. Work in VS Code Dev Container
3. Run build/test commands that take 15+ seconds
4. Receive completion notifications on Windows desktop

**Example commands in Dev Container:**
```bash
# These will trigger notifications if they take 15+ seconds
npm install
npm run build
yarn test
go test ./...
docker build .
```

### 7. Concurrent Operations

**Test multiple simultaneous operations:**
```bash
# Start daemon
./cmdbell --daemon start &

# Run multiple containers simultaneously
docker exec container1 sleep 16 &
docker exec container2 sleep 18 &
docker exec container3 sleep 20 &

# Also test local commands
./cmdbell sleep 17 &
```

**Expected:** Multiple notifications should appear as each command completes.

### 8. Error Handling and Edge Cases

**Docker not available:**
```bash
sudo systemctl stop docker  # Linux
./cmdbell --daemon start    # Should fail gracefully
```

**Duplicate daemon:**
```bash
./cmdbell --daemon start &
./cmdbell --daemon start    # Should detect already running
```

**Invalid commands:**
```bash
./cmdbell nonexistent-command  # Should handle gracefully
```

**Container cleanup during execution:**
```bash
docker run --rm ubuntu:20.04 sleep 30 &
# Kill container while sleep is running
docker kill $(docker ps -q | head -1)
```

### 9. Log Inspection

**Daemon logs:**
```bash
# Linux/macOS
tail -f ~/.cmdbell.log

# Windows
Get-Content $env:USERPROFILE\.cmdbell.log -Wait
```

**PID file:**
```bash
# Linux/macOS
cat ~/.cmdbell.pid

# Windows
Get-Content $env:USERPROFILE\.cmdbell.pid
```

## Expected Behavior

### Notifications Should Include:
- ✅ Command name and arguments
- ✅ Execution duration (rounded to seconds)  
- ✅ Success/failure status with appropriate icons
- ✅ Container name (for container commands)
- ✅ Timestamp information

### Platform-Specific Notifications:
- **Linux:** `notify-send` desktop notifications (if GUI available)
- **macOS:** Native notification center alerts
- **Windows:** Toast notifications in system tray area

### Console Fallback:
All platforms show console output regardless of GUI availability:
```
🔔 CmdBell: Command 'sleep' completed after 16s
🔔 CmdBell - Container: Command 'sleep 20' in 'test-container' completed after 20s
```

## Troubleshooting

### No notifications appearing:
1. Check daemon is running: `./cmdbell --daemon status`
2. Verify Docker connection: `docker ps`
3. Check notification settings on OS
4. Review logs for errors

### Permission issues:
- Linux: Ensure user is in docker group
- Windows: Run PowerShell as administrator if needed

### Container not detected:
- Ensure Docker daemon is accessible
- Check container is actually running: `docker ps`
- Verify command duration exceeds 15 seconds

## Performance Notes
- Minimal CPU/memory usage when idle
- Scales to monitor hundreds of concurrent containers
- No impact on monitored command performance
- Automatic cleanup of completed operations