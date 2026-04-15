package tools

import (
	"context"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

func TestListNamespaces_Success(t *testing.T) {
	orig := GetClientSet
	defer func() { GetClientSet = orig }()

	client := fake.NewSimpleClientset(
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "ns-alpha"}},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "ns-beta"}},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "ns-gamma"}},
	)
	GetClientSet = func() (kubernetes.Interface, error) {
		return client, nil
	}

	_, out, err := ListNamespaces(context.Background(), nil, struct{}{})
	if err != nil {
		t.Fatalf("ListNamespaces returned error: %v", err)
	}

	for _, name := range []string{"ns-alpha", "ns-beta", "ns-gamma"} {
		if !strings.Contains(out.Namespaces, name) {
			t.Errorf("expected %s in output, got: %q", name, out.Namespaces)
		}
	}
}

func TestListNamespaces_Empty(t *testing.T) {
	orig := GetClientSet
	defer func() { GetClientSet = orig }()

	client := fake.NewSimpleClientset()
	GetClientSet = func() (kubernetes.Interface, error) {
		return client, nil
	}

	_, out, err := ListNamespaces(context.Background(), nil, struct{}{})
	if err != nil {
		t.Fatalf("ListNamespaces returned error: %v", err)
	}
	if out.Namespaces != "" {
		t.Errorf("expected empty output for no namespaces, got: %q", out.Namespaces)
	}
}

func TestGetAllNamespaces_Success(t *testing.T) {
	orig := GetClientSet
	defer func() { GetClientSet = orig }()

	client := fake.NewSimpleClientset(
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "default"}},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "kube-system"}},
	)
	GetClientSet = func() (kubernetes.Interface, error) {
		return client, nil
	}

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
