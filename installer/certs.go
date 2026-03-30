package main

import (
	_ "embed"
	"os"
	"os/exec"
)

//go:embed assets/csr.conf
var csrConf []byte

func generateCerts() error {
	// Write csr.conf
	if err := os.WriteFile("/home/root/reDirector/csr.conf", csrConf, 0644); err != nil {
		return err
	}

	// Generate CA key and cert
	if err := exec.Command("openssl", "genrsa", "-out", "/home/root/reDirector/ca.key", "2048").Run(); err != nil {
		return err
	}
	if err := exec.Command("openssl", "req", "-new", "-sha256", "-x509", "-key", "/home/root/reDirector/ca.key", "-out", "/home/root/reDirector/ca.crt", "-days", "3650", "-subj", "/CN=reDirector").Run(); err != nil {
		return err
	}

	// Generate proxy private and public keys
	if err := exec.Command("openssl", "genrsa", "-out", "/home/root/reDirector/proxy.key", "2048").Run(); err != nil {
		return err
	}
	if err := exec.Command("openssl", "rsa", "-in", "/home/root/reDirector/proxy.key", "-pubout", "-out", "/home/root/reDirector/proxy.pub").Run(); err != nil {
		return err
	}

	// Generate certificate
	if err := exec.Command("openssl", "req", "-new", "-config", "/home/root/reDirector/csr.conf", "-key", "/home/root/reDirector/proxy.key", "-out", "/home/root/reDirector/proxy.csr").Run(); err != nil {
		return err
	}
	return exec.Command("openssl", "x509", "-req", "-in", "/home/root/reDirector/proxy.csr", "-CA", "/home/root/reDirector/ca.crt", "-CAkey", "/home/root/reDirector/ca.key", "-CAcreateserial", "-out", "/home/root/reDirector/proxy.crt", "-days", "3650", "-extfile", "/home/root/reDirector/csr.conf", "-extensions", "caext").Run()
}

func installCerts() error {
	// Copy CA cert
	src, err := os.ReadFile("/home/root/reDirector/ca.crt")
	if err != nil {
		return err
	}

	err = os.MkdirAll("/usr/local/share/ca-certificates", 0755)
	if err != nil {
		return err
	}

	if err := os.WriteFile("/usr/local/share/ca-certificates/reDirector.crt", src, 0644); err != nil {
		return err
	}

	// Refetch CA cert
	return exec.Command("update-ca-certificates").Run()
}

func uninstallCerts() error {
	// Delete CA cert
	if err := os.Remove("/usr/local/share/ca-certificates/reDirector.crt"); err != nil {
		return err
	}

	// Refetch CA cert
	return exec.Command("update-ca-certificates").Run()
}
