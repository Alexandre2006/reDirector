package main

import (
	"bufio"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strings"
)

func install() {
	slog.Info("installing reDirector")

	// Get address
	address := promptAddress()

	// Enable R/W on filesystem
	if err := unlockFS(); err != nil {
		slog.Error("failed to unlock filesystem", "err", err)
		return
	}
	slog.Info("filesystem unlocked")

	// Write config, certs, hosts, and service
	if err := writeConfig(address); err != nil {
		slog.Error("failed to write config", "err", err)
		return
	}
	slog.Info("config written")
	if err := generateCerts(); err != nil {
		slog.Error("failed to generate certificates", "err", err)
		return
	}
	slog.Info("certificates generated")
	if err := installCerts(); err != nil {
		slog.Error("failed to install certificates", "err", err)
		return
	}
	slog.Info("certificates installed")
	if err := installHosts(); err != nil {
		slog.Error("failed to patch hosts", "err", err)
		return
	}
	slog.Info("hosts patched")
	if err := installService(); err != nil {
		slog.Error("failed to install service", "err", err)
		return
	}
	slog.Info("service installed")

	// Reset sync and restart xochitl
	if err := stopService("xochitl"); err != nil {
		slog.Error("failed to stop xochitl", "err", err)
		return
	}
	slog.Info("xochitl stopped")
	if err := resetSync(); err != nil {
		slog.Error("failed to reset sync", "err", err)
		return
	}
	slog.Info("sync reset")
	if err := startService("xochitl"); err != nil {
		slog.Error("failed to start xochitl", "err", err)
		return
	}
	slog.Info("xochitl started")

	slog.Info("installation complete")
}

func uninstall() {
	slog.Info("uninstalling reDirector")

	// Enable R/W on filesystem
	if err := unlockFS(); err != nil {
		slog.Error("failed to unlock filesystem", "err", err)
		return
	}
	slog.Info("filesystem unlocked")

	// Remove service, hosts, certs, and config (don't error if they don't exist)
	if err := uninstallService(); err != nil {
		slog.Error("failed to uninstall service", "err", err)
	} else {
		slog.Info("service uninstalled")
	}
	if err := uninstallHosts(); err != nil {
		slog.Error("failed to uninstall hosts", "err", err)
	} else {
		slog.Info("hosts uninstalled")
	}
	if err := uninstallCerts(); err != nil {
		slog.Error("failed to uninstall certificates", "err", err)
	} else {
		slog.Info("certificates uninstalled")
	}
	if err := deleteConfig(); err != nil {
		slog.Error("failed to delete config", "err", err)
	} else {
		slog.Info("config deleted")
	}

	// Reset sync and restart xochitl
	if err := stopService("xochitl"); err != nil {
		slog.Error("failed to stop xochitl", "err", err)
		return
	}
	slog.Info("xochitl stopped")
	if err := resetSync(); err != nil {
		slog.Error("failed to reset sync", "err", err)
		return
	}
	slog.Info("sync reset")
	if err := startService("xochitl"); err != nil {
		slog.Error("failed to start xochitl", "err", err)
		return
	}
	slog.Info("xochitl started")

	slog.Info("uninstallation complete")
}

func repair() {
	slog.Info("repairing reDirector")

	// Enable R/W on filesystem
	if err := unlockFS(); err != nil {
		slog.Error("failed to unlock filesystem", "err", err)
		return
	}
	slog.Info("filesystem unlocked")

	// Write certs, hosts, and service
	if err := installCerts(); err != nil {
		slog.Error("failed to install certificates", "err", err)
		return
	}
	slog.Info("certificates installed")
	if err := installHosts(); err != nil {
		slog.Error("failed to patch hosts", "err", err)
		return
	}
	slog.Info("hosts patched")
	if err := installService(); err != nil {
		slog.Error("failed to install service", "err", err)
		return
	}
	slog.Info("service installed")

	// Reset sync and restart xochitl
	if err := stopService("xochitl"); err != nil {
		slog.Error("failed to stop xochitl", "err", err)
		return
	}
	slog.Info("xochitl stopped")
	if err := resetSync(); err != nil {
		slog.Error("failed to reset sync", "err", err)
		return
	}
	slog.Info("sync reset")
	if err := startService("xochitl"); err != nil {
		slog.Error("failed to start xochitl", "err", err)
		return
	}
	slog.Info("xochitl started")

	slog.Info("repair complete")
}

func promptAddress() string {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("Enter address: ")

		address, err := reader.ReadString('\n')
		if err != nil {
			return ""
		}

		address = strings.TrimSpace(address)
		parsed, err := url.Parse(address)
		if err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Host != "" {
			return address
		}

		fmt.Println("Please enter a valid http or https URL.")
	}
}
