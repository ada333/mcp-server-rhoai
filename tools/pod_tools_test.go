package tools

import (
	"context"
	"testing"

	core "github.com/amaly/mcp-server-rhoai/core"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestListPods_Success(t *testing.T) {
	ns := "test-ns"
	setupFakeClientSet(t,
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

	_, out, err := ListPods(context.Background(), nil, core.ListWorkbenchesInput{Namespace: ns})
	if err != nil {
		t.Fatalf("ListPods returned error: %v", err)
	}
	if len(out.Pods) != 1 {
		t.Fatalf("expected 1 pod, got %d", len(out.Pods))
	}
	if out.Pods[0].Name != "pod-a" || out.Pods[0].Status != "Running" {
		t.Errorf("expected pod-a with Running status, got: %+v", out.Pods[0])
	}
}
