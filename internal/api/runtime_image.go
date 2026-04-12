package api

import (
	"os"
	"strings"
)

const envRuntimeImageRef = "RUNTIME_IMAGE_REF"

// defaultRuntimeImageRef is the canonical fallback runtime image reference used
// when callers do not provide an explicit image_ref and no environment override
// is configured.
const defaultRuntimeImageRef = "ghcr.io/socioprophet/cloudshell-fog/runtime:dev"

// resolveRuntimeImageRef returns the runtime image reference in descending order
// of precedence:
//  1. explicit API request image_ref
//  2. RUNTIME_IMAGE_REF environment override
//  3. canonical code default
func resolveRuntimeImageRef(explicit string) string {
	if v := strings.TrimSpace(explicit); v != "" {
		return v
	}
	if v := strings.TrimSpace(os.Getenv(envRuntimeImageRef)); v != "" {
		return v
	}
	return defaultRuntimeImageRef
}
