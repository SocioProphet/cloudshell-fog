// Package delivery validates policy-gated SCOPE-D delivery envelopes before
// they are accepted by CloudShell Fog as edge assurance work.
package delivery

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

// DeliveryEnvelope describes a non-executing SCOPE-D work package that may be
// staged for operator review at a CloudShell Fog edge node.
type DeliveryEnvelope struct {
	SchemaVersion            string   `json:"schemaVersion"`
	EnvelopeID               string   `json:"envelopeId"`
	SourceSystem             string   `json:"sourceSystem"`
	Purpose                  string   `json:"purpose"`
	ArtifactRefs             []string `json:"artifactRefs"`
	RequiredPolicyRefs       []string `json:"requiredPolicyRefs"`
	OperatorApprovalRequired bool     `json:"operatorApprovalRequired"`
	ExecutionAllowed         bool     `json:"executionAllowed"`
	ExecutionPerformed       bool     `json:"executionPerformed"`
	NetworkAccessAllowed     bool     `json:"networkAccessAllowed"`
	MutationAllowed          bool     `json:"mutationAllowed"`
	CredentialAccessAllowed  bool     `json:"credentialAccessAllowed"`
	PayloadDeliveryAllowed   bool     `json:"payloadDeliveryAllowed"`
	ReceiptHash              string   `json:"receiptHash"`
}

// Sentinel validation errors.
var (
	ErrUnsupportedSchema       = errors.New("unsupported delivery envelope schema")
	ErrInvalidEnvelopeID       = errors.New("invalid delivery envelope id")
	ErrInvalidSourceSystem     = errors.New("invalid source system")
	ErrInvalidPurpose          = errors.New("invalid delivery purpose")
	ErrMissingArtifacts        = errors.New("missing artifact references")
	ErrMissingPolicies         = errors.New("missing required policy references")
	ErrOperatorApprovalMissing = errors.New("operator approval requirement missing")
	ErrUnsafeCapability        = errors.New("unsafe delivery capability requested")
	ErrInvalidReceiptHash      = errors.New("invalid receipt hash")
)

// Validate enforces CloudShell Fog's default posture for SCOPE-D delivery:
// reviewable, receipt-backed, policy-gated, and non-executing.
func Validate(envelope DeliveryEnvelope) error {
	if envelope.SchemaVersion != "0.1.0" {
		return ErrUnsupportedSchema
	}
	if !strings.HasPrefix(envelope.EnvelopeID, "cloudshell-fog-delivery:") {
		return ErrInvalidEnvelopeID
	}
	if envelope.SourceSystem != "scope-d" {
		return ErrInvalidSourceSystem
	}
	if envelope.Purpose != "edge_assurance_review" && envelope.Purpose != "policy_gated_delivery_review" {
		return ErrInvalidPurpose
	}
	if len(envelope.ArtifactRefs) == 0 {
		return ErrMissingArtifacts
	}
	if len(envelope.RequiredPolicyRefs) == 0 {
		return ErrMissingPolicies
	}
	if !envelope.OperatorApprovalRequired {
		return ErrOperatorApprovalMissing
	}
	if envelope.ExecutionAllowed || envelope.ExecutionPerformed || envelope.NetworkAccessAllowed || envelope.MutationAllowed || envelope.CredentialAccessAllowed || envelope.PayloadDeliveryAllowed {
		return ErrUnsafeCapability
	}
	if !strings.HasPrefix(envelope.ReceiptHash, "sha256:") || len(strings.TrimPrefix(envelope.ReceiptHash, "sha256:")) != 64 {
		return ErrInvalidReceiptHash
	}
	return nil
}

// ComputeReceiptHash returns a deterministic hash over the policy-significant
// fields. It intentionally excludes ReceiptHash itself.
func ComputeReceiptHash(envelope DeliveryEnvelope) string {
	material := strings.Join([]string{
		envelope.SchemaVersion,
		envelope.EnvelopeID,
		envelope.SourceSystem,
		envelope.Purpose,
		strings.Join(envelope.ArtifactRefs, ","),
		strings.Join(envelope.RequiredPolicyRefs, ","),
		fmt.Sprintf("approval=%t", envelope.OperatorApprovalRequired),
		fmt.Sprintf("executionAllowed=%t", envelope.ExecutionAllowed),
		fmt.Sprintf("networkAccessAllowed=%t", envelope.NetworkAccessAllowed),
		fmt.Sprintf("mutationAllowed=%t", envelope.MutationAllowed),
		fmt.Sprintf("credentialAccessAllowed=%t", envelope.CredentialAccessAllowed),
		fmt.Sprintf("payloadDeliveryAllowed=%t", envelope.PayloadDeliveryAllowed),
	}, "|")
	sum := sha256.Sum256([]byte(material))
	return "sha256:" + hex.EncodeToString(sum[:])
}
