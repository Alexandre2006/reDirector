package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

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

func readConfig(path string) (*Config, *url.URL, error) {
	var cfg Config

	b, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("read config: %w", err)
	}

	if err := toml.Unmarshal(b, &cfg); err != nil {
		return nil, nil, fmt.Errorf("parse config: %w", err)
	}

	upstream, err := validateConfig(&cfg)
	if err != nil {
		return nil, nil, err
	}

	return &cfg, upstream, nil
}

func validateConfig(cfg *Config) (*url.URL, error) {
	cfg.CertificatePath = strings.TrimSpace(cfg.CertificatePath)
	cfg.CertificateKeyPath = strings.TrimSpace(cfg.CertificateKeyPath)
	cfg.ClientCertificatePath = strings.TrimSpace(cfg.ClientCertificatePath)
	cfg.ClientCertificateKeyPath = strings.TrimSpace(cfg.ClientCertificateKeyPath)
	cfg.ListenAddress = strings.TrimSpace(cfg.ListenAddress)
	cfg.UpstreamAddress = strings.TrimSpace(cfg.UpstreamAddress)

	switch {
	case cfg.CertificatePath == "":
		return nil, errors.New("missing CertificatePath")
	case cfg.CertificateKeyPath == "":
		return nil, errors.New("missing CertificateKeyPath")
	case cfg.ListenAddress == "":
		return nil, errors.New("missing ListenAddress")
	case cfg.UpstreamAddress == "":
		return nil, errors.New("missing UpstreamAddress")
	}

	clientCertSet := cfg.ClientCertificatePath != ""
	clientKeySet := cfg.ClientCertificateKeyPath != ""

	switch {
	case clientCertSet && !clientKeySet:
		return nil, errors.New("ClientCertificatePath is set but ClientCertificateKeyPath is missing")
	case !clientCertSet && clientKeySet:
		return nil, errors.New("ClientCertificateKeyPath is set but ClientCertificatePath is missing")
	}

	upstream, err := url.Parse(cfg.UpstreamAddress)
	if err != nil {
		return nil, fmt.Errorf("parse UpstreamAddress: %w", err)
	}

	if upstream.Scheme == "" {
		return nil, errors.New("UpstreamAddress is missing scheme")
	}
	if upstream.Host == "" {
		return nil, errors.New("UpstreamAddress is missing host")
	}

	switch upstream.Scheme {
	case "http", "https":
	default:
		return nil, fmt.Errorf("unsupported upstream scheme %q", upstream.Scheme)
	}

	return upstream, nil
}

func buildClientTLSConfig(cfg *Config) (*tls.Config, error) {
	if cfg.ClientCertificatePath == "" && cfg.ClientCertificateKeyPath == "" {
		slog.Info("client certificate disabled")
		return nil, nil
	}

	clientCert, err := tls.LoadX509KeyPair(cfg.ClientCertificatePath, cfg.ClientCertificateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("load client certificate: %w", err)
	}

	slog.Info("client certificate enabled")

	return &tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{clientCert},
	}, nil
}

func main() {
	cfg, upstream, err := readConfig(configPath)
	if err != nil {
		slog.Error("failed to read config", "path", configPath, "err", err)
		os.Exit(1)
	}

	clientTLSConfig, err := buildClientTLSConfig(cfg)
	if err != nil {
		slog.Error("failed to configure client TLS", "err", err)
		os.Exit(1)
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil
	transport.TLSClientConfig = clientTLSConfig

	proxy := &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetURL(upstream)
			r.SetXForwarded()
		},
		Transport: transport,
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			slog.Error("proxy request failed", "method", r.Method, "url", r.URL.String(), "err", err)
			http.Error(w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)
		},
		ErrorLog: log.New(os.Stderr, "reverseproxy: ", log.LstdFlags),
	}

	srv := &http.Server{
		Addr:              cfg.ListenAddress,
		Handler:           proxy,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
		ErrorLog: log.New(os.Stderr, "http-server: ", log.LstdFlags),
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		slog.Info("redirector starting", "addr", cfg.ListenAddress, "upstream", upstream.String())
		errCh <- srv.ListenAndServeTLS(cfg.CertificatePath, cfg.CertificateKeyPath)
	}()

	select {
	case <-ctx.Done():
		slog.Info("shutdown signal received")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			slog.Error("graceful shutdown failed", "err", err)
		}

		err := <-errCh
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server exited with error", "err", err)
			os.Exit(1)
		}

		slog.Info("server stopped")

	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server failed", "err", err)
			os.Exit(1)
		}
	}
}
