package api

import (
	"os"
	"testing"
)

func TestResolveRuntimeImageRef_ExplicitWins(t *testing.T) {
	if got := resolveRuntimeImageRef("custom/image:tag"); got != "custom/image:tag" {
		t.Fatalf("expected explicit image to win, got %q", got)
	}
}

func TestResolveRuntimeImageRef_EnvFallback(t *testing.T) {
	old := os.Getenv(envRuntimeImageRef)
	defer os.Setenv(envRuntimeImageRef, old)
	if err := os.Setenv(envRuntimeImageRef, "env/image:tag"); err != nil {
		t.Fatalf("set env: %v", err)
	}

	if got := resolveRuntimeImageRef(""); got != "env/image:tag" {
		t.Fatalf("expected env image, got %q", got)
	}
}

func TestResolveRuntimeImageRef_Default(t *testing.T) {
	old := os.Getenv(envRuntimeImageRef)
	defer os.Setenv(envRuntimeImageRef, old)
	_ = os.Unsetenv(envRuntimeImageRef)

	if got := resolveRuntimeImageRef(""); got != defaultRuntimeImageRef {
		t.Fatalf("expected default image, got %q", got)
	}
}
