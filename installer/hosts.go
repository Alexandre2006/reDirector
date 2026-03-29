package main

import (
	_ "embed"
	"errors"
	"os"
	"strings"
)

//go:embed assets/hosts
var hosts []byte

func installHosts() error {
	// Remove patch if it already exists
	if err := uninstallHosts(); err != nil {
		return err
	}

	// Append content to hosts file
	file, err := os.OpenFile("/etc/hosts", os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}

	if _, err := file.Write(hosts); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()

}

func uninstallHosts() error {
	// Read existing content
	current, err := os.ReadFile("/etc/hosts")
	if err != nil {
		return err
	}
	content := string(current)

	// Trim content
	start := strings.Index(content, "# reDirector_Start")
	if start == -1 {
		return nil
	}

	end := strings.Index(content[start:], "# reDirector_End")
	if end == -1 {
		return errors.New("found reDirector_Start but not reDirector_End in /etc/hosts")
	}
	end += start + len("# reDirector_End")

	updated := strings.TrimSpace(content[:start] + content[end:])
	if updated != "" {
		updated += "\n"
	}

	// Write updated content
	return os.WriteFile("/etc/hosts", []byte(updated), 0o644)
}
