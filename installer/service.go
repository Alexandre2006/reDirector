package main

import (
	_ "embed"
	"os"
)

//go:embed assets/reDirector.service
var service []byte

//go:embed assets/reDirector
var executable []byte

func installService() error {
	// Write executable file
	if err := os.WriteFile("/home/root/reDirector/reDirector", executable, 0755); err != nil {
		return err
	}

	// Write service file
	if err := os.WriteFile("/etc/systemd/system/reDirector.service", service, 0644); err != nil {
		return err
	}

	// Reload systemd, enable service, and start it
	if err := daemonReload(); err != nil {
		return err
	}
	if err := enableService("reDirector"); err != nil {
		return err
	}
	return startService("reDirector")
}

func uninstallService() error {
	// Stop and disable service
	if err := stopService("reDirector"); err != nil {
		return err
	}
	if err := disableService("reDirector"); err != nil {
		return err
	}

	// Delete service file
	return os.Remove("/etc/systemd/system/reDirector.service")
}
