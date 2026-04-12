package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/SocioProphet/cloudshell-fog/internal/connector"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	envConnectorMode = "CONNECTOR_MODE"
	envKubeconfig    = "KUBECONFIG"
)

// connectorModeFromEnv resolves the requested connector mode.
//
// Supported values:
// - stub (default for local/dev)
// - k8s  (requires a valid Kubernetes REST config)
func connectorModeFromEnv() string {
	mode := strings.TrimSpace(strings.ToLower(os.Getenv(envConnectorMode)))
	if mode == "" {
		return "stub"
	}
	return mode
}

// buildConnector resolves the runtime connector from environment.
//
// This helper is intentionally kept separate from run() so the gateway entrypoint
// can fail fast on unsupported or misconfigured connector modes.
func buildConnector(logger *slog.Logger) (connector.Connector, string, error) {
	switch mode := connectorModeFromEnv(); mode {
	case "stub":
		logger.Info("using stub connector", "mode", mode)
		return connector.NewStubConnector(), mode, nil
	case "k8s":
		cfg, source, err := loadKubernetesRESTConfig()
		if err != nil {
			return nil, mode, fmt.Errorf("load kubernetes config: %w", err)
		}
		conn, err := connector.NewKubernetesConnector(cfg, "cloudshell-")
		if err != nil {
			return nil, mode, fmt.Errorf("create kubernetes connector: %w", err)
		}
		logger.Info("using kubernetes connector", "mode", mode, "config_source", source)
		return conn, mode, nil
	default:
		return nil, mode, fmt.Errorf("unsupported connector mode %q (expected stub or k8s)", mode)
	}
}

// loadKubernetesRESTConfig resolves Kubernetes client configuration.
//
// Resolution order:
// 1. KUBECONFIG path, if set
// 2. in-cluster config (ServiceAccount / projected credentials)
func loadKubernetesRESTConfig() (*rest.Config, string, error) {
	if kubeconfig := strings.TrimSpace(os.Getenv(envKubeconfig)); kubeconfig != "" {
		cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, "kubeconfig", err
		}
		return cfg, "kubeconfig", nil
	}
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, "in-cluster", err
	}
	return cfg, "in-cluster", nil
}
