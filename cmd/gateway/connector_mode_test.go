package main

import (
	"os"
	"testing"
)

func TestConnectorModeFromEnv_DefaultsToStub(t *testing.T) {
	old := os.Getenv(envConnectorMode)
	defer os.Setenv(envConnectorMode, old)
	_ = os.Unsetenv(envConnectorMode)

	if got := connectorModeFromEnv(); got != "stub" {
		t.Fatalf("expected default connector mode stub, got %q", got)
	}
}

func TestConnectorModeFromEnv_NormalizesCaseAndWhitespace(t *testing.T) {
	old := os.Getenv(envConnectorMode)
	defer os.Setenv(envConnectorMode, old)
	if err := os.Setenv(envConnectorMode, "  K8S  "); err != nil {
		t.Fatalf("set env: %v", err)
	}

	if got := connectorModeFromEnv(); got != "k8s" {
		t.Fatalf("expected normalized connector mode k8s, got %q", got)
	}
}

func TestConnectorModeFromEnv_PreservesUnknownValuesForValidation(t *testing.T) {
	old := os.Getenv(envConnectorMode)
	defer os.Setenv(envConnectorMode, old)
	if err := os.Setenv(envConnectorMode, "weird-mode"); err != nil {
		t.Fatalf("set env: %v", err)
	}

	if got := connectorModeFromEnv(); got != "weird-mode" {
		t.Fatalf("expected unknown connector mode to pass through for validation, got %q", got)
	}
}
