// here should be unit tests for the tools
// in the tests we use mocking to inject fake clientset and dynamic client

package main

import (
	"context"
	"strings"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

func TestListPods_Success(t *testing.T) {
	orig := getClientSet
	defer func() { getClientSet = orig }()

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
	getClientSet = func() (kubernetes.Interface, error) {
		return client, nil
	}

	_, out, err := ListPods(context.Background(), nil, ListWorkbenchesInput{Namespace: ns})
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

func newUnstructuredWorkbench(name, namespace string) *unstructured.Unstructured {
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(workbenchesGVR.GroupVersion().WithKind("Notebook"))
	u.SetName(name)
	u.SetNamespace(namespace)
	return u
}

func TestListWorkbenches(t *testing.T) {
	orig := getDynamicClient
	defer func() { getDynamicClient = orig }()

	ns := "test-ns"
	scheme := runtime.NewScheme()
	client := dynamicfake.NewSimpleDynamicClient(scheme,
		newUnstructuredWorkbench("wb-1", ns),
		newUnstructuredWorkbench("wb-other", "other-ns"),
	)

	getDynamicClient = func() (dynamic.Interface, error) {
		return client, nil
	}

	_, out, err := ListWorkbenches(context.Background(), nil, ListWorkbenchesInput{Namespace: ns})
	if err != nil {
		t.Fatalf("ListWorkbenches returned error: %v", err)
	}

	if !strings.Contains(out.Workbenches, "- wb-1\n") {
		t.Errorf("expected wb-1 in output, got: %q", out.Workbenches)
	}
	if strings.Contains(out.Workbenches, "wb-other") {
		t.Errorf("did not expect wb-other in output, got: %q", out.Workbenches)
	}
}

func TestListAllWorkbenches(t *testing.T) {
	orig := getDynamicClient
	defer func() { getDynamicClient = orig }()

	scheme := runtime.NewScheme()
	client := dynamicfake.NewSimpleDynamicClient(scheme,
		newUnstructuredWorkbench("wb-1", "ns1"),
		newUnstructuredWorkbench("wb-2", "ns2"),
	)

	getDynamicClient = func() (dynamic.Interface, error) {
		return client, nil
	}

	_, out, err := ListAllWorkbenches(context.Background(), nil, ListWorkbenchesInput{Namespace: ""})
	if err != nil {
		t.Fatalf("ListAllWorkbenches returned error: %v", err)
	}

	if !strings.Contains(out.Workbenches, "- wb-1\n") {
		t.Errorf("expected wb-1 in output, got: %q", out.Workbenches)
	}
	if !strings.Contains(out.Workbenches, "- wb-2\n") {
		t.Errorf("expected wb-2 in output, got: %q", out.Workbenches)
	}
}

// TODO
func TestChangeWorkbenchStatus(t *testing.T) {
	orig := getDynamicClient
	defer func() { getDynamicClient = orig }()

	scheme := runtime.NewScheme()
	stoppedWorkbench := newUnstructuredWorkbench("StoppedWorkbench", "ns1")
	stoppedWorkbench.SetAnnotations(map[string]string{
		"kubeflow-resource-stopped": time.Now().UTC().Format(time.RFC3339),
	})
	runningWorkbench := newUnstructuredWorkbench("RunningWorkbench", "ns1")

	client := dynamicfake.NewSimpleDynamicClient(scheme,
		stoppedWorkbench,
		runningWorkbench,
	)

	getDynamicClient = func() (dynamic.Interface, error) {
		return client, nil
	}

	// tests all the combinations of status changes
	// maybe checking the annotations would be better than output message
	_, out, err := ChangeWorkbenchStatus(context.Background(), nil, ChangeWorkbenchStatusInput{Namespace: "ns1", WorkbenchName: "StoppedWorkbench", Status: Stopped})
	if err != nil {
		t.Fatalf("ChangeWorkbenchStatus returned error: %v", err)
	}
	if out.Message != "Workbench StoppedWorkbench is already stopped" {
		t.Errorf("expected StoppedWorkbench is already stopped, got: %q", out.Message)
	}

	_, out, err = ChangeWorkbenchStatus(context.Background(), nil, ChangeWorkbenchStatusInput{Namespace: "ns1", WorkbenchName: "RunningWorkbench", Status: Running})
	if err != nil {
		t.Fatalf("ChangeWorkbenchStatus returned error: %v", err)
	}
	if out.Message != "Workbench RunningWorkbench is already running" {
		t.Errorf("expected RunningWorkbench is already running, got: %q", out.Message)
	}

	_, out, err = ChangeWorkbenchStatus(context.Background(), nil, ChangeWorkbenchStatusInput{Namespace: "ns1", WorkbenchName: "StoppedWorkbench", Status: Running})
	if err != nil {
		t.Fatalf("ChangeWorkbenchStatus returned error: %v", err)
	}
	if out.Message != "Workbench StoppedWorkbench is running" {
		t.Errorf("expected StoppedWorkbench is running, got: %q", out.Message)
	}

	_, out, err = ChangeWorkbenchStatus(context.Background(), nil, ChangeWorkbenchStatusInput{Namespace: "ns1", WorkbenchName: "RunningWorkbench", Status: Stopped})
	if err != nil {
		t.Fatalf("ChangeWorkbenchStatus returned error: %v", err)
	}
	if out.Message != "Workbench RunningWorkbench is stopped" {
		t.Errorf("expected RunningWorkbench is stopped, got: %q", out.Message)
	}
}

// TODO
func TestCreateWorkbench(t *testing.T) {
}

// TODO
func TestListImages(t *testing.T) {
}
