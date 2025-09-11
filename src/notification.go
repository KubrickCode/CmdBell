package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

func sendNotification(command string, duration time.Duration, success bool) {
	status := "completed"
	icon := "✅"
	if !success {
		status = "failed"
		icon = "❌"
	}

	title := "CmdBell"
	message := fmt.Sprintf("Command '%s' %s after %s",
		command, status, duration.Round(time.Second))

	// Always show console output as fallback
	fmt.Printf("\n🔔 %s: %s\n", title, message)

	// Send native OS notification
	err := sendNativeNotification(title, message, icon)
	if err != nil {
		fmt.Printf("Failed to send native notification: %v\n", err)
	}
}

func sendContainerNotification(command, containerName string, duration time.Duration, success bool) {
	status := "completed"
	icon := "✅"
	if !success {
		status = "failed"
		icon = "❌"
	}

	title := "CmdBell - Container"
	message := fmt.Sprintf("Command '%s' in '%s' %s after %s",
		command, containerName, status, duration.Round(time.Second))

	// Always show console output as fallback
	fmt.Printf("\n🔔 %s: %s\n", title, message)

	// Send native OS notification
	err := sendNativeNotification(title, message, icon)
	if err != nil {
		fmt.Printf("Failed to send native notification: %v\n", err)
	}
}

func sendNativeNotification(title, message, icon string) error {
	switch runtime.GOOS {
	case "darwin":
		return sendMacOSNotification(title, message, icon)
	case "linux":
		return sendLinuxNotification(title, message, icon)
	case "windows":
		return sendWindowsNotification(title, message, icon)
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

func sendMacOSNotification(title, message, icon string) error {
	script := fmt.Sprintf(`display notification "%s" with title "%s" subtitle "%s"`,
		escapeAppleScript(message), escapeAppleScript(title), icon)

	cmd := exec.Command("osascript", "-e", script)
	return cmd.Run()
}

func sendLinuxNotification(title, message, icon string) error {
	// Check if we're in a headless environment
	if os.Getenv("DISPLAY") == "" && os.Getenv("WAYLAND_DISPLAY") == "" {
		return fmt.Errorf("no GUI environment detected (headless mode)")
	}

	// Try notify-send first (most common)
	if _, err := exec.LookPath("notify-send"); err == nil {
		cmd := exec.Command("notify-send", title, message, "--icon=info")
		if err := cmd.Run(); err == nil {
			return nil
		}
	}

	// Fallback to kdialog (KDE)
	if _, err := exec.LookPath("kdialog"); err == nil {
		cmd := exec.Command("kdialog", "--passivepopup", fmt.Sprintf("%s\n%s", title, message), "5")
		if err := cmd.Run(); err == nil {
			return nil
		}
	}

	// Fallback to zenity (GNOME)
	if _, err := exec.LookPath("zenity"); err == nil {
		cmd := exec.Command("zenity", "--info", "--text", fmt.Sprintf("%s\n%s", title, message), "--timeout=5")
		if err := cmd.Run(); err == nil {
			return nil
		}
	}

	return fmt.Errorf("no working notification tool found or GUI not available")
}

func sendWindowsNotification(title, message, icon string) error {
	// Use PowerShell to show Windows toast notification
	script := fmt.Sprintf(`
		Add-Type -AssemblyName System.Windows.Forms;
		$balloon = New-Object System.Windows.Forms.NotifyIcon;
		$balloon.Icon = [System.Drawing.SystemIcons]::Information;
		$balloon.BalloonTipIcon = "Info";
		$balloon.BalloonTipText = "%s";
		$balloon.BalloonTipTitle = "%s";
		$balloon.Visible = $true;
		$balloon.ShowBalloonTip(5000);
		Start-Sleep -Seconds 6;
		$balloon.Dispose();
	`, escapeWindowsString(message), escapeWindowsString(title))

	cmd := exec.Command("powershell", "-Command", script)
	return cmd.Run()
}

func escapeAppleScript(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	return s
}

func escapeWindowsString(s string) string {
	s = strings.ReplaceAll(s, "\"", "\\\"")
	return s
}