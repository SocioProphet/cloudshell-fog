package connector

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestApplySessionNetworkPoliciesCreatesExpectedPolicies(t *testing.T) {
	ctx := context.Background()
	client := fake.NewSimpleClientset()
	connector := &KubernetesConnector{client: client, nsPrefix: "cloudshell-"}

	namespace := "cloudshell-sess-123"
	sessionID := "sess-123"

	if err := connector.applySessionNetworkPolicies(ctx, namespace, sessionID); err != nil {
		t.Fatalf("apply network policies: %v", err)
	}

	policies, err := client.NetworkingV1().NetworkPolicies(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		t.Fatalf("list network policies: %v", err)
	}
	if len(policies.Items) != 3 {
		t.Fatalf("expected 3 network policies, got %d", len(policies.Items))
	}

	byName := map[string]bool{}
	for _, policy := range policies.Items {
		byName[policy.Name] = true
		if policy.Namespace != namespace {
			t.Fatalf("policy %s created in namespace %q, want %q", policy.Name, policy.Namespace, namespace)
		}
	}

	for _, name := range []string{"default-deny-all", "allow-gateway-to-shell", "allow-shell-egress"} {
		if !byName[name] {
			t.Fatalf("missing policy %s; got %#v", name, byName)
		}
	}

	defaultDeny, err := client.NetworkingV1().NetworkPolicies(namespace).Get(ctx, "default-deny-all", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("get default-deny policy: %v", err)
	}
	if len(defaultDeny.Spec.PolicyTypes) != 2 {
		t.Fatalf("expected ingress+egress default deny, got %#v", defaultDeny.Spec.PolicyTypes)
	}

	gatewayPolicy, err := client.NetworkingV1().NetworkPolicies(namespace).Get(ctx, "allow-gateway-to-shell", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("get gateway policy: %v", err)
	}
	if gatewayPolicy.Spec.PodSelector.MatchLabels["cloudshell.io/session-id"] != sessionID {
		t.Fatalf("gateway policy session selector mismatch: %#v", gatewayPolicy.Spec.PodSelector.MatchLabels)
	}
	if gatewayPolicy.Spec.PodSelector.MatchLabels["cloudshell.io/managed"] != "true" {
		t.Fatalf("gateway policy managed selector mismatch: %#v", gatewayPolicy.Spec.PodSelector.MatchLabels)
	}
	if len(gatewayPolicy.Spec.Ingress) != 1 || len(gatewayPolicy.Spec.Ingress[0].From) != 1 {
		t.Fatalf("expected one gateway ingress peer, got %#v", gatewayPolicy.Spec.Ingress)
	}

	egressPolicy, err := client.NetworkingV1().NetworkPolicies(namespace).Get(ctx, "allow-shell-egress", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("get egress policy: %v", err)
	}
	if egressPolicy.Spec.PodSelector.MatchLabels["cloudshell.io/session-id"] != sessionID {
		t.Fatalf("egress policy session selector mismatch: %#v", egressPolicy.Spec.PodSelector.MatchLabels)
	}
	if len(egressPolicy.Spec.Egress) != 2 {
		t.Fatalf("expected DNS and HTTPS egress rules, got %#v", egressPolicy.Spec.Egress)
	}
}
