package tools

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestListNamespaces_Success(t *testing.T) {
	setupFakeClientSet(t,
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "ns-alpha"}},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "ns-beta"}},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "ns-gamma"}},
	)

	_, out, err := ListNamespaces(context.Background(), nil, struct{}{})
	if err != nil {
		t.Fatalf("ListNamespaces returned error: %v", err)
	}

	if len(out.Namespaces) != 3 {
		t.Fatalf("expected 3 namespaces, got %d", len(out.Namespaces))
	}

	expected := map[string]bool{"ns-alpha": true, "ns-beta": true, "ns-gamma": true}
	for _, ns := range out.Namespaces {
		if !expected[ns] {
			t.Errorf("unexpected namespace %s in output", ns)
		}
	}
}

func TestListNamespaces_Empty(t *testing.T) {
	setupFakeClientSet(t)

	_, out, err := ListNamespaces(context.Background(), nil, struct{}{})
	if err != nil {
		t.Fatalf("ListNamespaces returned error: %v", err)
	}
	if len(out.Namespaces) != 0 {
		t.Errorf("expected empty output for no namespaces, got: %v", out.Namespaces)
	}
}

func TestGetAllNamespaces_Success(t *testing.T) {
	setupFakeClientSet(t,
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "default"}},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "kube-system"}},
	)

	names, err := GetAllNamespaces(context.Background())
	if err != nil {
		t.Fatalf("GetAllNamespaces returned error: %v", err)
	}
	if len(names) != 2 {
		t.Fatalf("expected 2 namespaces, got %d", len(names))
	}
	if names[0] != "default" || names[1] != "kube-system" {
		t.Errorf("expected [default, kube-system], got: %v", names)
	}
}
