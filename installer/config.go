package main

import (
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

const configPath = "/home/root/reDirector/config.toml"

type Config struct {
	CertificatePath          string
	CertificateKeyPath       string
	ClientCertificatePath    string
	ClientCertificateKeyPath string
	ListenAddress            string
	UpstreamAddress          string
}

var defaultConfig = Config{
	CertificatePath:          "proxy.crt",
	CertificateKeyPath:       "proxy.key",
	ClientCertificatePath:    "",
	ClientCertificateKeyPath: "",
	ListenAddress:            ":443",
	UpstreamAddress:          "",
}

func writeConfig(upstreamAddress string) error {
	cfg := defaultConfig
	cfg.UpstreamAddress = upstreamAddress

	data, err := toml.Marshal(cfg)
	if err != nil {
		return err
	}

	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		return err
	}

	return nil
}

func deleteConfig() error {
	return os.RemoveAll(filepath.Dir(configPath))
}
