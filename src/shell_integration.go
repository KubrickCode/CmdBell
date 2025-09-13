package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ShellIntegration struct {
	executablePath string
	homeDir        string
}

func NewShellIntegration() (*ShellIntegration, error) {
	executablePath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to get executable path: %v", err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %v", err)
	}

	return &ShellIntegration{
		executablePath: executablePath,
		homeDir:        homeDir,
	}, nil
}

func (si *ShellIntegration) Install() error {
	shells := []string{"bash", "zsh", "fish"}

	fmt.Println("🔧 Installing CmdBell shell integration...")

	for _, shell := range shells {
		if err := si.installForShell(shell); err != nil {
			fmt.Printf("⚠️  Warning: Failed to install for %s: %v\n", shell, err)
		} else {
			fmt.Printf("✅ Installed for %s\n", shell)
		}
	}

	fmt.Println("\n🎉 Shell integration installed!")
	fmt.Println("💡 Restart your shell or run 'source ~/.bashrc' (or equivalent) to activate")
	return nil
}

func (si *ShellIntegration) Uninstall() error {
	shells := []string{"bash", "zsh", "fish"}

	fmt.Println("🗑️  Removing CmdBell shell integration...")

	for _, shell := range shells {
		if err := si.uninstallForShell(shell); err != nil {
			fmt.Printf("⚠️  Warning: Failed to remove from %s: %v\n", shell, err)
		} else {
			fmt.Printf("✅ Removed from %s\n", shell)
		}
	}

	fmt.Println("🎉 Shell integration removed!")
	return nil
}

func (si *ShellIntegration) installForShell(shell string) error {
	switch shell {
	case "bash":
		return si.installBash()
	case "zsh":
		return si.installZsh()
	case "fish":
		return si.installFish()
	default:
		return fmt.Errorf("unsupported shell: %s", shell)
	}
}

func (si *ShellIntegration) uninstallForShell(shell string) error {
	switch shell {
	case "bash":
		return si.uninstallBash()
	case "zsh":
		return si.uninstallZsh()
	case "fish":
		return si.uninstallFish()
	default:
		return fmt.Errorf("unsupported shell: %s", shell)
	}
}

func (si *ShellIntegration) installBash() error {
	bashrcPath := filepath.Join(si.homeDir, ".bashrc")

	bashHook := si.generateBashHook()
	return si.addToShellConfig(bashrcPath, bashHook)
}

func (si *ShellIntegration) installZsh() error {
	zshrcPath := filepath.Join(si.homeDir, ".zshrc")

	zshHook := si.generateZshHook()
	return si.addToShellConfig(zshrcPath, zshHook)
}

func (si *ShellIntegration) installFish() error {
	fishConfigDir := filepath.Join(si.homeDir, ".config", "fish", "config.fish")

	// Create fish config directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(fishConfigDir), 0755); err != nil {
		return fmt.Errorf("failed to create fish config directory: %v", err)
	}

	fishHook := si.generateFishHook()
	return si.addToShellConfig(fishConfigDir, fishHook)
}

func (si *ShellIntegration) generateBashHook() string {
	return `
# CmdBell shell integration - START
_cmdbell_preexec() {
    export CMDBELL_START_TIME=$(date +%%s.%%N)
    export CMDBELL_COMMAND="$1"
}

_cmdbell_precmd() {
    if [[ -n "$CMDBELL_START_TIME" ]] && [[ -n "$CMDBELL_COMMAND" ]]; then
        local end_time=$(date +%%s.%%N)
        local duration=$(echo "$end_time - $CMDBELL_START_TIME" | bc -l)
        local duration_int=$(printf "%%.0f" "$duration")
        
        if [[ $duration_int -ge 15 ]]; then
            local exit_code=$?
            local success="true"
            [[ $exit_code -ne 0 ]] && success="false"
            
            # Try to detect Docker host IP
            local host_ip="localhost"
            if [[ -f "/.dockerenv" ]] || [[ -n "$DOCKER_HOST" ]]; then
                # Running in container, try Docker host IPs
                if command -v nslookup >/dev/null 2>&1; then
                    if nslookup host.docker.internal >/dev/null 2>&1; then
                        host_ip="host.docker.internal"
                    elif nslookup docker.for.windows.localhost >/dev/null 2>&1; then
                        host_ip="docker.for.windows.localhost"
                    elif nslookup docker.for.mac.localhost >/dev/null 2>&1; then
                        host_ip="docker.for.mac.localhost"
                    fi
                fi
            fi
            
            # Send HTTP notification
            local payload='{"command":"'"$CMDBELL_COMMAND"'","container_name":"'"${HOSTNAME:-unknown}"'","duration":"'"${duration_int}s"'","success":'"$success"'}'
            
            # Try HTTP first, fallback to local notification
            if ! curl -s -X POST "http://$host_ip:59721/notify" \
                -H "Content-Type: application/json" \
                -d "$payload" >/dev/null 2>&1; then
                # HTTP failed, try local fallback if cmdbell binary exists
                if command -v cmdbell >/dev/null 2>&1; then
                    cmdbell --notify "$CMDBELL_COMMAND" "$duration_int" "$exit_code" &
                fi
            fi
        fi
        
        unset CMDBELL_START_TIME
        unset CMDBELL_COMMAND
    fi
}

# Set up hooks for bash
if [[ -n "$PS1" ]]; then
    trap '_cmdbell_preexec "$BASH_COMMAND"' DEBUG
    PROMPT_COMMAND="_cmdbell_precmd${PROMPT_COMMAND:+; $PROMPT_COMMAND}"
fi
# CmdBell shell integration - END
`
}

func (si *ShellIntegration) generateZshHook() string {
	return `
# CmdBell shell integration - START
_cmdbell_preexec() {
    export CMDBELL_START_TIME=$(date +%%s.%%N)
    export CMDBELL_COMMAND="$1"
}

_cmdbell_precmd() {
    if [[ -n "$CMDBELL_START_TIME" ]] && [[ -n "$CMDBELL_COMMAND" ]]; then
        local end_time=$(date +%%s.%%N)
        local duration=$(echo "$end_time - $CMDBELL_START_TIME" | bc -l 2>/dev/null || echo "0")
        local duration_int=$(printf "%%.0f" "$duration")
        
        if [[ $duration_int -ge 15 ]]; then
            local exit_code=$?
            local success="true"
            [[ $exit_code -ne 0 ]] && success="false"
            
            # Try to detect Docker host IP
            local host_ip="localhost"
            if [[ -f "/.dockerenv" ]] || [[ -n "$DOCKER_HOST" ]]; then
                # Running in container, try Docker host IPs
                if command -v nslookup >/dev/null 2>&1; then
                    if nslookup host.docker.internal >/dev/null 2>&1; then
                        host_ip="host.docker.internal"
                    elif nslookup docker.for.windows.localhost >/dev/null 2>&1; then
                        host_ip="docker.for.windows.localhost"
                    elif nslookup docker.for.mac.localhost >/dev/null 2>&1; then
                        host_ip="docker.for.mac.localhost"
                    fi
                fi
            fi
            
            # Send HTTP notification
            local payload='{"command":"'"$CMDBELL_COMMAND"'","container_name":"'"${HOSTNAME:-unknown}"'","duration":"'"${duration_int}s"'","success":'"$success"'}'
            
            # Try HTTP first, fallback to local notification
            if ! curl -s -X POST "http://$host_ip:59721/notify" \
                -H "Content-Type: application/json" \
                -d "$payload" >/dev/null 2>&1; then
                # HTTP failed, try local fallback if cmdbell binary exists
                if command -v cmdbell >/dev/null 2>&1; then
                    cmdbell --notify "$CMDBELL_COMMAND" "$duration_int" "$exit_code" &
                fi
            fi
        fi
        
        unset CMDBELL_START_TIME
        unset CMDBELL_COMMAND
    fi
}

# Set up hooks for zsh
if [[ -n "$PS1" ]]; then
    autoload -Uz add-zsh-hook
    add-zsh-hook preexec _cmdbell_preexec
    add-zsh-hook precmd _cmdbell_precmd
fi
# CmdBell shell integration - END
`
}

func (si *ShellIntegration) generateFishHook() string {
	return `
# CmdBell shell integration - START
function _cmdbell_preexec --on-event fish_preexec
    set -gx CMDBELL_START_TIME (date +%%s.%%N)
    set -gx CMDBELL_COMMAND "$argv"
end

function _cmdbell_postcmd --on-event fish_postexec
    if test -n "$CMDBELL_START_TIME"; and test -n "$CMDBELL_COMMAND"
        set end_time (date +%%s.%%N)
        set duration (math "$end_time - $CMDBELL_START_TIME")
        set duration_int (printf "%%.0f" "$duration")
        
        if test $duration_int -ge 15
            set exit_code $status
            set success "true"
            if test $exit_code -ne 0
                set success "false"
            end
            
            # Try to detect Docker host IP
            set host_ip "localhost"
            if test -f "/.dockerenv"; or test -n "$DOCKER_HOST"
                # Running in container, try Docker host IPs
                if command -v nslookup >/dev/null 2>&1
                    if nslookup host.docker.internal >/dev/null 2>&1
                        set host_ip "host.docker.internal"
                    else if nslookup docker.for.windows.localhost >/dev/null 2>&1
                        set host_ip "docker.for.windows.localhost"
                    else if nslookup docker.for.mac.localhost >/dev/null 2>&1
                        set host_ip "docker.for.mac.localhost"
                    end
                end
            end
            
            # Send HTTP notification
            set payload '{"command":"'"$CMDBELL_COMMAND"'","container_name":"'(hostname)'","duration":"'"$duration_int"'s","success":'"$success"'}'
            
            # Try HTTP first, fallback to local notification
            if not curl -s -X POST "http://$host_ip:59721/notify" \
                -H "Content-Type: application/json" \
                -d "$payload" >/dev/null 2>&1
                # HTTP failed, try local fallback if cmdbell binary exists
                if command -v cmdbell >/dev/null 2>&1
                    cmdbell --notify "$CMDBELL_COMMAND" "$duration_int" "$exit_code" &
                end
            end
        end
        
        set -e CMDBELL_START_TIME
        set -e CMDBELL_COMMAND
    end
end
# CmdBell shell integration - END
`
}

func (si *ShellIntegration) addToShellConfig(configPath, hookContent string) error {
	startMarker := "# CmdBell shell integration - START"
	endMarker := "# CmdBell shell integration - END"

	// Read existing config
	var existingContent string
	if content, err := os.ReadFile(configPath); err == nil {
		existingContent = string(content)
	}

	// Remove existing hook if present
	cleanContent := si.removeExistingHook(existingContent, startMarker, endMarker)

	// Add new hook
	newContent := cleanContent + "\n" + hookContent + "\n"

	return os.WriteFile(configPath, []byte(newContent), 0644)
}

func (si *ShellIntegration) removeExistingHook(content, startMarker, endMarker string) string {
	startIdx := strings.Index(content, startMarker)
	if startIdx == -1 {
		return content
	}

	endIdx := strings.Index(content[startIdx:], endMarker)
	if endIdx == -1 {
		return content
	}

	endIdx += startIdx + len(endMarker)

	// Remove the hook section (including trailing newlines)
	before := strings.TrimRight(content[:startIdx], "\n")
	after := strings.TrimLeft(content[endIdx:], "\n")

	if before == "" {
		return after
	}
	if after == "" {
		return before
	}
	return before + "\n" + after
}

func (si *ShellIntegration) uninstallBash() error {
	bashrcPath := filepath.Join(si.homeDir, ".bashrc")
	return si.removeFromShellConfig(bashrcPath)
}

func (si *ShellIntegration) uninstallZsh() error {
	zshrcPath := filepath.Join(si.homeDir, ".zshrc")
	return si.removeFromShellConfig(zshrcPath)
}

func (si *ShellIntegration) uninstallFish() error {
	fishConfigPath := filepath.Join(si.homeDir, ".config", "fish", "config.fish")
	return si.removeFromShellConfig(fishConfigPath)
}

func (si *ShellIntegration) removeFromShellConfig(configPath string) error {
	startMarker := "# CmdBell shell integration - START"
	endMarker := "# CmdBell shell integration - END"

	content, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist, nothing to remove
		}
		return fmt.Errorf("failed to read config file: %v", err)
	}

	cleanContent := si.removeExistingHook(string(content), startMarker, endMarker)

	return os.WriteFile(configPath, []byte(cleanContent), 0644)
}
