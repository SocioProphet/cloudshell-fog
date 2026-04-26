package connector

import (
	"context"
	"fmt"
	"io"
	"time"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

// KubernetesConnector implements Connector using the Kubernetes API.
// Each session gets its own namespace (nsPrefix+sessionID) for isolation.
type KubernetesConnector struct {
	client     kubernetes.Interface
	restConfig *rest.Config
	nsPrefix   string // e.g. "cloudshell-"
}

// NewKubernetesConnector creates a KubernetesConnector from an existing rest.Config.
func NewKubernetesConnector(cfg *rest.Config, nsPrefix string) (*KubernetesConnector, error) {
	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("create k8s client: %w", err)
	}
	return &KubernetesConnector{client: client, restConfig: cfg, nsPrefix: nsPrefix}, nil
}

func (k *KubernetesConnector) nsName(sessionID string) string {
	return k.nsPrefix + sessionID
}

// Allocate creates a per-session namespace and a shell Pod with the given profile.
func (k *KubernetesConnector) Allocate(ctx context.Context, sessionID string, profile Profile, placement string, imageRef string) (RuntimeRef, error) {
	ns := k.nsName(sessionID)

	_, err := k.client.CoreV1().Namespaces().Create(ctx, &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: ns,
			Labels: map[string]string{
				"cloudshell.io/session-id": sessionID,
				"cloudshell.io/managed":    "true",
			},
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return RuntimeRef{}, fmt.Errorf("create namespace: %w", err)
	}

	if err := k.applySessionNetworkPolicies(ctx, ns, sessionID); err != nil {
		return RuntimeRef{}, fmt.Errorf("apply session network policies: %w", err)
	}

	cpuQty, err := resource.ParseQuantity(profile.CPU)
	if err != nil {
		return RuntimeRef{}, fmt.Errorf("parse cpu quantity %q: %w", profile.CPU, err)
	}
	memQty, err := resource.ParseQuantity(profile.Memory)
	if err != nil {
		return RuntimeRef{}, fmt.Errorf("parse memory quantity %q: %w", profile.Memory, err)
	}

	const podName = "shell"
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: ns,
			Labels: map[string]string{
				"cloudshell.io/session-id": sessionID,
				"cloudshell.io/managed":    "true",
			},
		},
		Spec: corev1.PodSpec{
			RestartPolicy:                corev1.RestartPolicyNever,
			AutomountServiceAccountToken: boolPtr(false),
			Containers: []corev1.Container{{
				Name:    "shell",
				Image:   imageRef,
				Command: []string{"/bin/bash"},
				Stdin:   true,
				TTY:     true,
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    cpuQty,
						corev1.ResourceMemory: memQty,
					},
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    cpuQty,
						corev1.ResourceMemory: memQty,
					},
				},
				SecurityContext: &corev1.SecurityContext{
					AllowPrivilegeEscalation: boolPtr(false),
					RunAsNonRoot:             boolPtr(true),
					SeccompProfile: &corev1.SeccompProfile{
						Type: corev1.SeccompProfileTypeRuntimeDefault,
					},
					Capabilities: &corev1.Capabilities{
						Drop: []corev1.Capability{"ALL"},
					},
				},
			}},
		},
	}

	if _, err := k.client.CoreV1().Pods(ns).Create(ctx, pod, metav1.CreateOptions{}); err != nil {
		return RuntimeRef{}, fmt.Errorf("create pod: %w", err)
	}

	if err := k.waitForPodRunning(ctx, ns, podName); err != nil {
		return RuntimeRef{}, fmt.Errorf("pod did not reach Running state: %w", err)
	}

	return RuntimeRef{
		ID:        sessionID,
		Namespace: ns,
		PodName:   podName,
		NodeID:    placement,
	}, nil
}

func (k *KubernetesConnector) applySessionNetworkPolicies(ctx context.Context, namespace, sessionID string) error {
	policies := []*networkingv1.NetworkPolicy{
		defaultDenyAllNetworkPolicy(namespace),
		allowGatewayIngressNetworkPolicy(namespace, sessionID),
		allowShellEgressNetworkPolicy(namespace, sessionID),
	}
	for _, policy := range policies {
		if _, err := k.client.NetworkingV1().NetworkPolicies(namespace).Create(ctx, policy, metav1.CreateOptions{}); err != nil {
			return fmt.Errorf("create networkpolicy %s: %w", policy.Name, err)
		}
	}
	return nil
}

func defaultDenyAllNetworkPolicy(namespace string) *networkingv1.NetworkPolicy {
	return &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{Name: "default-deny-all", Namespace: namespace},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{},
			PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeIngress, networkingv1.PolicyTypeEgress},
		},
	}
}

func allowGatewayIngressNetworkPolicy(namespace, sessionID string) *networkingv1.NetworkPolicy {
	return &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{Name: "allow-gateway-to-shell", Namespace: namespace},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{MatchLabels: managedSessionLabels(sessionID)},
			PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeIngress},
			Ingress: []networkingv1.NetworkPolicyIngressRule{{
				From: []networkingv1.NetworkPolicyPeer{{
					NamespaceSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"kubernetes.io/metadata.name": "cloudshell-system"}},
					PodSelector:       &metav1.LabelSelector{MatchLabels: map[string]string{"app": "cloudshell-gateway"}},
				}},
			}},
		},
	}
}

func allowShellEgressNetworkPolicy(namespace, sessionID string) *networkingv1.NetworkPolicy {
	return &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{Name: "allow-shell-egress", Namespace: namespace},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{MatchLabels: managedSessionLabels(sessionID)},
			PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeEgress},
			Egress: []networkingv1.NetworkPolicyEgressRule{
				{Ports: []networkingv1.NetworkPolicyPort{{Protocol: protocolPtr(corev1.ProtocolUDP), Port: intstrPtr(53)}, {Protocol: protocolPtr(corev1.ProtocolTCP), Port: intstrPtr(53)}}},
				{Ports: []networkingv1.NetworkPolicyPort{{Protocol: protocolPtr(corev1.ProtocolTCP), Port: intstrPtr(443)}}},
			},
		},
	}
}

func managedSessionLabels(sessionID string) map[string]string {
	return map[string]string{
		"cloudshell.io/session-id": sessionID,
		"cloudshell.io/managed":    "true",
	}
}

func protocolPtr(protocol corev1.Protocol) *corev1.Protocol { return &protocol }

func intstrPtr(port int) *intstr.IntOrString {
	v := intstr.FromInt(port)
	return &v
}

func (k *KubernetesConnector) waitForPodRunning(ctx context.Context, namespace, podName string) error {
	return wait.PollUntilContextTimeout(ctx, 2*time.Second, 2*time.Minute, true,
		func(ctx context.Context) (bool, error) {
			pod, err := k.client.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
			if err != nil {
				return false, err
			}
			switch pod.Status.Phase {
			case corev1.PodRunning:
				return true, nil
			case corev1.PodFailed, corev1.PodSucceeded:
				return false, fmt.Errorf("pod entered terminal phase: %s", pod.Status.Phase)
			default:
				return false, nil
			}
		})
}

// AttachPTY opens an exec session on the shell container and returns PTY streams.
func (k *KubernetesConnector) AttachPTY(ctx context.Context, ref RuntimeRef) (*PTYStreams, error) {
	req := k.client.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(ref.PodName).
		Namespace(ref.Namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: "shell",
			Command:   []string{"/bin/bash"},
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       true,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(k.restConfig, "POST", req.URL())
	if err != nil {
		return nil, fmt.Errorf("create SPDY executor: %w", err)
	}

	stdinR, stdinW := io.Pipe()
	stdoutR, stdoutW := io.Pipe()
	sq := newTerminalSizeQueue()

	go func() {
		_ = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
			Stdin:             stdinR,
			Stdout:            stdoutW,
			Stderr:            stdoutW,
			Tty:               true,
			TerminalSizeQueue: sq,
		})
		stdoutW.Close()
	}()

	return &PTYStreams{
		Stdin:  stdinW,
		Stdout: stdoutR,
		Resize: func(cols, rows uint16) error {
			sq.push(remotecommand.TerminalSize{Width: cols, Height: rows})
			return nil
		},
	}, nil
}

// EnforceLimits is a no-op; limits are enforced via Pod ResourceRequirements.
// Extend to query the metrics-server API for real-time enforcement.
func (k *KubernetesConnector) EnforceLimits(_ context.Context, _ RuntimeRef) error {
	return nil
}

// Terminate deletes the per-session namespace, cascading to all contained resources.
func (k *KubernetesConnector) Terminate(ctx context.Context, ref RuntimeRef) error {
	return k.client.CoreV1().Namespaces().Delete(ctx, ref.Namespace, metav1.DeleteOptions{})
}

// terminalSizeQueue implements remotecommand.TerminalSizeQueue via a buffered channel.
type terminalSizeQueue struct {
	ch chan remotecommand.TerminalSize
}

func newTerminalSizeQueue() *terminalSizeQueue {
	return &terminalSizeQueue{ch: make(chan remotecommand.TerminalSize, 8)}
}

func (t *terminalSizeQueue) Next() *remotecommand.TerminalSize {
	size, ok := <-t.ch
	if !ok {
		return nil
	}
	return &size
}

func (t *terminalSizeQueue) push(s remotecommand.TerminalSize) {
	select {
	case t.ch <- s:
	default: // drop if queue full to avoid blocking the resize caller
	}
}

func boolPtr(b bool) *bool { return &b }
