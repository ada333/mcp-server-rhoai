package tools

import (
	"context"
	"strings"
	"testing"

	core "github.com/amaly/mcp-server-rhoai/core"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

func TestListPods_Success(t *testing.T) {
	orig := GetClientSet
	defer func() { GetClientSet = orig }()

	ns := "test-ns"
	client := fake.NewSimpleClientset(
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod-a",
				Namespace: ns,
			},
			Status: corev1.PodStatus{Phase: corev1.PodRunning},
		},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod-other",
				Namespace: "other-ns",
			},
			Status: corev1.PodStatus{Phase: corev1.PodSucceeded},
		},
	)
	GetClientSet = func() (kubernetes.Interface, error) {
		return client, nil
	}

	_, out, err := ListPods(context.Background(), nil, core.ListWorkbenchesInput{Namespace: ns})
	if err != nil {
		t.Fatalf("ListPods returned error: %v", err)
	}
	if !strings.Contains(out.Pods, "- pod-a (Running)\n") {
		t.Errorf("expected pod-a Running in output, got: %q", out.Pods)
	}
	if strings.Contains(out.Pods, "pod-other") {
		t.Errorf("did not expect pod-other in output, got: %q", out.Pods)
	}
}
