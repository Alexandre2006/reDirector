package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func unlockFS() error {
	// Get device ID
	file, err := os.ReadFile("/sys/devices/soc0/machine")
	if err != nil {
		return err
	}
	deviceID := string(file)
	deviceID = strings.TrimSpace(deviceID)

	// Check if device needs unlocking
	if deviceID == "reMarkable 1.0" || deviceID == "reMarkable Prototype 1" || deviceID == "reMarkable 2.0" {
		return nil
	} else if deviceID == "reMarkable Ferrari" || deviceID == "reMarkable Chiappa" {
		if err := exec.Command("mount", "-o", "remount,rw", "/").Run(); err != nil {
			return err
		}
		out, err := exec.Command("umount", "-R", "/etc").CombinedOutput()
		if err != nil && !strings.Contains(string(out), "not mounted") {
			return err
		}
		return nil
	}

	return errors.New("unsupported device ID")
}

func resetSync() error {
	matches, err := filepath.Glob("/home/root/.local/share/remarkable/xochitl/*.metadata")
	if err != nil {
		return err
	}

	for _, path := range matches {
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		updated := strings.ReplaceAll(string(data), `synced": true`, `synced": false`)
		if updated == string(data) {
			continue
		}

		if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
			return err
		}
	}

	return nil
}

func startService(serviceName string) error {
	return exec.Command("systemctl", "start", serviceName).Run()
}

func stopService(serviceName string) error {
	return exec.Command("systemctl", "stop", serviceName).Run()
}

func enableService(serviceName string) error {
	return exec.Command("systemctl", "enable", serviceName).Run()
}

func disableService(serviceName string) error {
	return exec.Command("systemctl", "disable", serviceName).Run()
}

func daemonReload() error {
	return exec.Command("systemctl", "daemon-reload").Run()
}
