package delivery

import (
	"errors"
	"testing"
)

func validEnvelope() DeliveryEnvelope {
	envelope := DeliveryEnvelope{
		SchemaVersion:            "0.1.0",
		EnvelopeID:               "cloudshell-fog-delivery:scope-d-demo-001",
		SourceSystem:             "scope-d",
		Purpose:                  "edge_assurance_review",
		ArtifactRefs:             []string{"scope-d://cyber-graph-export/demo", "scope-d://detection-candidate-export/demo"},
		RequiredPolicyRefs:       []string{"policyfabric://approval/operator-review-required"},
		OperatorApprovalRequired: true,
		ExecutionAllowed:         false,
		ExecutionPerformed:       false,
		NetworkAccessAllowed:     false,
		MutationAllowed:          false,
		CredentialAccessAllowed:  false,
		PayloadDeliveryAllowed:   false,
	}
	envelope.ReceiptHash = ComputeReceiptHash(envelope)
	return envelope
}

func TestValidateAcceptsSafeScopeDEnvelope(t *testing.T) {
	if err := Validate(validEnvelope()); err != nil {
		t.Fatalf("expected valid envelope, got %v", err)
	}
}

func TestValidateRejectsExecutionAllowed(t *testing.T) {
	envelope := validEnvelope()
	envelope.ExecutionAllowed = true
	envelope.ReceiptHash = ComputeReceiptHash(envelope)
	if err := Validate(envelope); !errors.Is(err, ErrUnsafeCapability) {
		t.Fatalf("expected unsafe capability error, got %v", err)
	}
}

func TestValidateRejectsNetworkMutationCredentialAndPayloadCapabilities(t *testing.T) {
	cases := map[string]func(*DeliveryEnvelope){
		"network":    func(e *DeliveryEnvelope) { e.NetworkAccessAllowed = true },
		"mutation":   func(e *DeliveryEnvelope) { e.MutationAllowed = true },
		"credential": func(e *DeliveryEnvelope) { e.CredentialAccessAllowed = true },
		"payload":    func(e *DeliveryEnvelope) { e.PayloadDeliveryAllowed = true },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			envelope := validEnvelope()
			mutate(&envelope)
			envelope.ReceiptHash = ComputeReceiptHash(envelope)
			if err := Validate(envelope); !errors.Is(err, ErrUnsafeCapability) {
				t.Fatalf("expected unsafe capability error, got %v", err)
			}
		})
	}
}

func TestValidateRequiresOperatorApproval(t *testing.T) {
	envelope := validEnvelope()
	envelope.OperatorApprovalRequired = false
	envelope.ReceiptHash = ComputeReceiptHash(envelope)
	if err := Validate(envelope); !errors.Is(err, ErrOperatorApprovalMissing) {
		t.Fatalf("expected operator approval error, got %v", err)
	}
}

func TestValidateRequiresPolicyRefs(t *testing.T) {
	envelope := validEnvelope()
	envelope.RequiredPolicyRefs = nil
	envelope.ReceiptHash = ComputeReceiptHash(envelope)
	if err := Validate(envelope); !errors.Is(err, ErrMissingPolicies) {
		t.Fatalf("expected missing policies error, got %v", err)
	}
}

func TestValidateRequiresReceiptHashShape(t *testing.T) {
	envelope := validEnvelope()
	envelope.ReceiptHash = "sha256:not-a-real-hash"
	if err := Validate(envelope); !errors.Is(err, ErrInvalidReceiptHash) {
		t.Fatalf("expected invalid receipt hash error, got %v", err)
	}
}
