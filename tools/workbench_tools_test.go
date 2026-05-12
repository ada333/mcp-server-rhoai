package tools

import (
	"context"
	"strings"
	"testing"
	"time"

	core "github.com/amaly/mcp-server-rhoai/core"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
)

func newUnstructuredWorkbench(name, namespace string) *unstructured.Unstructured {
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(core.WorkbenchesGVR.GroupVersion().WithKind("Notebook"))
	u.SetName(name)
	u.SetNamespace(namespace)
	return u
}

func newDetailedWorkbench(name, namespace, user, imageDisplayName, imageTag, hwProfile, pvcName string, stopped bool) *unstructured.Unstructured {
	annotations := map[string]interface{}{
		"opendatahub.io/username":                       user,
		"opendatahub.io/image-display-name":             imageDisplayName,
		"opendatahub.io/hardware-profile-name":          hwProfile,
		"notebooks.opendatahub.io/last-image-selection": imageDisplayName + ":" + imageTag,
	}
	if stopped {
		annotations["kubeflow-resource-stopped"] = time.Now().UTC().Format(time.RFC3339)
	}

	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "kubeflow.org/v1",
			"kind":       "Notebook",
			"metadata": map[string]interface{}{
				"name":        name,
				"namespace":   namespace,
				"annotations": annotations,
			},
			"spec": map[string]interface{}{
				"template": map[string]interface{}{
					"spec": map[string]interface{}{
						"containers": []interface{}{
							map[string]interface{}{
								"name":  name,
								"image": "quay.io/test:latest",
								"resources": map[string]interface{}{
									"requests": map[string]interface{}{
										"cpu":    "2",
										"memory": "4Gi",
									},
								},
							},
						},
						"volumes": []interface{}{
							map[string]interface{}{
								"name": pvcName,
								"persistentVolumeClaim": map[string]interface{}{
									"claimName": pvcName,
								},
							},
						},
					},
				},
			},
		},
	}
}

func mockWorkbenchHelpers(t *testing.T) {
	origUptime := getUptimeFromWorkbenchFn
	origDisk := getDiskUsageFromPVCFn
	origImageInfo := getImageInfoFn
	t.Cleanup(func() {
		getUptimeFromWorkbenchFn = origUptime
		getDiskUsageFromPVCFn = origDisk
		getImageInfoFn = origImageInfo
	})
	getUptimeFromWorkbenchFn = func(name, namespace string) (string, error) {
		return "2h30m", nil
	}
	getDiskUsageFromPVCFn = func(ctx context.Context, dyn dynamic.Interface, namespace, pvcName string) (string, error) {
		return "5Gi", nil
	}
	getImageInfoFn = func(ctx context.Context, displayName, version string) (string, string, string, error) {
		return "quay.io/modh/" + displayName, "abc123", displayName, nil
	}
}

func TestListWorkbenches_Stopped(t *testing.T) {
	mockWorkbenchHelpers(t)

	ns := "test-ns"
	wb := newDetailedWorkbench("wb-1", ns, "alice", "PyTorch", "v1", "small", "wb-1-pvc", true)
	setupFakeClient(t, wb)

	_, out, err := ListWorkbenches(context.Background(), nil, core.ListWorkbenchesInput{Namespace: ns})
	if err != nil {
		t.Fatalf("ListWorkbenches returned error: %v", err)
	}
	if len(out.Workbenches) != 1 {
		t.Fatalf("expected 1 workbench, got %d", len(out.Workbenches))
	}
	if out.Workbenches[0].Name != "wb-1" {
		t.Errorf("expected name wb-1, got %q", out.Workbenches[0].Name)
	}
	if out.Workbenches[0].Status != "stopped" {
		t.Errorf("expected status stopped, got %q", out.Workbenches[0].Status)
	}
	if out.Workbenches[0].User != "alice" {
		t.Errorf("expected user alice, got %q", out.Workbenches[0].User)
	}
	if out.Workbenches[0].CPUUsage != "2" {
		t.Errorf("expected CPU 2, got %q", out.Workbenches[0].CPUUsage)
	}
	if out.Workbenches[0].MemoryUsage != "4Gi" {
		t.Errorf("expected memory 4Gi, got %q", out.Workbenches[0].MemoryUsage)
	}
	if out.Workbenches[0].Uptime != "0s" {
		t.Errorf("expected uptime 0s for stopped workbench, got %q", out.Workbenches[0].Uptime)
	}
	if out.Workbenches[0].DiskUsage != "5Gi" {
		t.Errorf("expected disk 5Gi from mock, got %q", out.Workbenches[0].DiskUsage)
	}
}

func TestListWorkbenches_Running(t *testing.T) {
	mockWorkbenchHelpers(t)

	ns := "test-ns"
	wb := newDetailedWorkbench("running-wb", ns, "bob", "TensorFlow", "v2", "large", "running-pvc", false)
	setupFakeClient(t, wb)

	_, out, err := ListWorkbenches(context.Background(), nil, core.ListWorkbenchesInput{Namespace: ns})
	if err != nil {
		t.Fatalf("ListWorkbenches returned error: %v", err)
	}
	if len(out.Workbenches) != 1 {
		t.Fatalf("expected 1 workbench, got %d", len(out.Workbenches))
	}
	if out.Workbenches[0].Status != "running" {
		t.Errorf("expected status running, got %q", out.Workbenches[0].Status)
	}
	if out.Workbenches[0].Uptime != "2h30m" {
		t.Errorf("expected uptime 2h30m from mock, got %q", out.Workbenches[0].Uptime)
	}
}

func TestListWorkbenchesAllNamespaces(t *testing.T) {
	mockWorkbenchHelpers(t)

	wb1 := newDetailedWorkbench("wb-1", "ns1", "alice", "PyTorch", "v1", "small", "pvc1", true)
	wb2 := newDetailedWorkbench("wb-2", "ns2", "bob", "TF", "v1", "medium", "pvc2", true)
	setupFakeClient(t, wb1, wb2)

	_, out, err := ListWorkbenches(context.Background(), nil, core.ListWorkbenchesInput{})
	if err != nil {
		t.Fatalf("ListWorkbenches (all namespaces) returned error: %v", err)
	}
	if len(out.Workbenches) != 2 {
		t.Fatalf("expected 2 workbenches across all namespaces, got %d", len(out.Workbenches))
	}

	names := map[string]bool{}
	for _, wb := range out.Workbenches {
		names[wb.Name] = true
	}
	if !names["wb-1"] || !names["wb-2"] {
		t.Errorf("expected wb-1 and wb-2, got: %v", names)
	}
}

func TestDeleteWorkbench_Success(t *testing.T) {
	wb := newUnstructuredWorkbench("to-delete", "ns1")
	setupFakeClient(t, wb)

	_, out, err := DeleteWorkbench(context.Background(), nil, core.DeleteWorkbenchInput{
		Namespace:     "ns1",
		WorkbenchName: "to-delete",
	})
	if err != nil {
		t.Fatalf("DeleteWorkbench returned error: %v", err)
	}
	if !strings.Contains(out.Message, "successfully deleted") {
		t.Errorf("unexpected message: %q", out.Message)
	}
}

func TestDeleteWorkbench_NotFound(t *testing.T) {
	setupFakeClient(t)

	_, _, err := DeleteWorkbench(context.Background(), nil, core.DeleteWorkbenchInput{
		Namespace:     "ns1",
		WorkbenchName: "nonexistent",
	})
	if err == nil {
		t.Fatal("expected error deleting nonexistent workbench, got nil")
	}
}

func TestIsWorkbenchStopped(t *testing.T) {
	stoppedWb := newUnstructuredWorkbench("stopped-wb", "ns1")
	stoppedWb.SetAnnotations(map[string]string{
		"kubeflow-resource-stopped": time.Now().UTC().Format(time.RFC3339),
	})
	runningWb := newUnstructuredWorkbench("running-wb", "ns1")

	dyn := setupFakeClient(t, stoppedWb, runningWb)

	stopped, err := IsWorkbenchStopped(context.Background(), dyn, "ns1", "stopped-wb")
	if err != nil {
		t.Fatalf("IsWorkbenchStopped returned error: %v", err)
	}
	if !stopped {
		t.Error("expected stopped-wb to be stopped")
	}

	stopped, err = IsWorkbenchStopped(context.Background(), dyn, "ns1", "running-wb")
	if err != nil {
		t.Fatalf("IsWorkbenchStopped returned error: %v", err)
	}
	if stopped {
		t.Error("expected running-wb to not be stopped")
	}
}

func TestResolveWorkbenchStatus(t *testing.T) {
	tests := []struct {
		name        string
		annotations map[string]string
		want        string
	}{
		{"stopped", map[string]string{"kubeflow-resource-stopped": "2025-01-01T00:00:00Z"}, "stopped"},
		{"running", map[string]string{}, "running"},
		{"nil annotations", nil, "running"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			annotations := tt.annotations
			if annotations == nil {
				annotations = map[string]string{}
			}
			got := resolveWorkbenchStatus(annotations)
			if got != tt.want {
				t.Errorf("resolveWorkbenchStatus(%v) = %q, want %q", tt.annotations, got, tt.want)
			}
		})
	}
}

func TestParseImageTag(t *testing.T) {
	tests := []struct {
		name        string
		annotations map[string]string
		want        string
	}{
		{"valid tag", map[string]string{"notebooks.opendatahub.io/last-image-selection": "pytorch:v2024.1"}, "v2024.1"},
		{"no tag", map[string]string{"notebooks.opendatahub.io/last-image-selection": "pytorch"}, ""},
		{"empty", map[string]string{}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseImageTag(tt.annotations)
			if got != tt.want {
				t.Errorf("parseImageTag(%v) = %q, want %q", tt.annotations, got, tt.want)
			}
		})
	}
}

func TestBuildResourceRequirements(t *testing.T) {
	profile := core.HardwareProfile{
		HardwareProfileName: "test",
		Resources: []core.HardwareProfileResource{
			{ResourceIdentifier: "cpu", DefaultCount: "2", MaxCount: "8"},
			{ResourceIdentifier: "memory", DefaultCount: "4Gi", MaxCount: "16Gi"},
		},
	}
	limits, requests := buildResourceRequirements(profile)
	if limits["cpu"] != "8" {
		t.Errorf("expected cpu limit 8, got %v", limits["cpu"])
	}
	if requests["cpu"] != "2" {
		t.Errorf("expected cpu request 2, got %v", requests["cpu"])
	}
	if limits["memory"] != "16Gi" {
		t.Errorf("expected memory limit 16Gi, got %v", limits["memory"])
	}
	if requests["memory"] != "4Gi" {
		t.Errorf("expected memory request 4Gi, got %v", requests["memory"])
	}
}

func TestResolveFullImageURL(t *testing.T) {
	tests := []struct {
		repo, tag, want string
	}{
		{"quay.io/modh/pytorch", "v1", "quay.io/modh/pytorch:v1"},
		{"quay.io/modh/pytorch", "", "quay.io/modh/pytorch"},
	}
	for _, tt := range tests {
		got := resolveFullImageURL(tt.repo, tt.tag)
		if got != tt.want {
			t.Errorf("resolveFullImageURL(%q, %q) = %q, want %q", tt.repo, tt.tag, got, tt.want)
		}
	}
}

func TestGetPVCNameFromWorkbench(t *testing.T) {
	wb := newDetailedWorkbench("wb", "ns", "user", "img", "v1", "hw", "my-pvc", true)
	pvcName, err := getPVCNameFromWorkbench(wb)
	if err != nil {
		t.Fatalf("getPVCNameFromWorkbench returned error: %v", err)
	}
	if pvcName != "my-pvc" {
		t.Errorf("expected my-pvc, got %q", pvcName)
	}
}

func TestGetPVCNameFromWorkbench_NoVolumes(t *testing.T) {
	wb := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "kubeflow.org/v1",
			"kind":       "Notebook",
			"metadata":   map[string]interface{}{"name": "wb", "namespace": "ns"},
			"spec": map[string]interface{}{
				"template": map[string]interface{}{
					"spec": map[string]interface{}{
						"containers": []interface{}{},
					},
				},
			},
		},
	}
	pvcName, err := getPVCNameFromWorkbench(wb)
	if err != nil {
		t.Fatalf("getPVCNameFromWorkbench returned error: %v", err)
	}
	if pvcName != "" {
		t.Errorf("expected empty PVC name, got %q", pvcName)
	}
}

func TestGetResourceRequestsFromWorkbench(t *testing.T) {
	wb := newDetailedWorkbench("wb", "ns", "user", "img", "v1", "hw", "pvc", true)
	cpu, mem, gpu, err := getResourceRequestsFromWorkbench(wb)
	if err != nil {
		t.Fatalf("getResourceRequestsFromWorkbench returned error: %v", err)
	}
	if cpu != "2" {
		t.Errorf("expected cpu=2, got %q", cpu)
	}
	if mem != "4Gi" {
		t.Errorf("expected mem=4Gi, got %q", mem)
	}
	if gpu != "" {
		t.Errorf("expected empty gpu, got %q", gpu)
	}
}

func TestChangeWorkbenchStatus(t *testing.T) {
	stoppedWorkbench := newUnstructuredWorkbench("StoppedWorkbench", "ns1")
	stoppedWorkbench.SetAnnotations(map[string]string{
		"kubeflow-resource-stopped": time.Now().UTC().Format(time.RFC3339),
	})
	runningWorkbench := newUnstructuredWorkbench("RunningWorkbench", "ns1")

	setupFakeClient(t, stoppedWorkbench, runningWorkbench)

	_, out, err := ChangeWorkbenchStatus(context.Background(), nil, core.ChangeWorkbenchStatusInput{Namespace: "ns1", WorkbenchName: "StoppedWorkbench", Status: core.Stopped})
	if err != nil {
		t.Fatalf("ChangeWorkbenchStatus returned error: %v", err)
	}
	if out.Message != "Workbench StoppedWorkbench is already stopped" {
		t.Errorf("expected StoppedWorkbench is already stopped, got: %q", out.Message)
	}

	_, out, err = ChangeWorkbenchStatus(context.Background(), nil, core.ChangeWorkbenchStatusInput{Namespace: "ns1", WorkbenchName: "RunningWorkbench", Status: core.Running})
	if err != nil {
		t.Fatalf("ChangeWorkbenchStatus returned error: %v", err)
	}
	if out.Message != "Workbench RunningWorkbench is already running" {
		t.Errorf("expected RunningWorkbench is already running, got: %q", out.Message)
	}

	_, out, err = ChangeWorkbenchStatus(context.Background(), nil, core.ChangeWorkbenchStatusInput{Namespace: "ns1", WorkbenchName: "StoppedWorkbench", Status: core.Running})
	if err != nil {
		t.Fatalf("ChangeWorkbenchStatus returned error: %v", err)
	}

	if out.Message != "Workbench StoppedWorkbench is running" {
		t.Errorf("expected StoppedWorkbench is running, got: %q", out.Message)
	}

	_, out, err = ChangeWorkbenchStatus(context.Background(), nil, core.ChangeWorkbenchStatusInput{Namespace: "ns1", WorkbenchName: "RunningWorkbench", Status: core.Stopped})
	if err != nil {
		t.Fatalf("ChangeWorkbenchStatus returned error: %v", err)
	}
	if out.Message != "Workbench RunningWorkbench is stopped" {
		t.Errorf("expected RunningWorkbench is stopped, got: %q", out.Message)
	}
}

func TestCreateWorkbench_Success(t *testing.T) {
	mockWorkbenchHelpers(t)

	pvc := newUnstructuredPVC("existing-pvc", "ns1", "10Gi")
	setupFakeClient(t, pvc)

	_, out, err := CreateWorkbench(context.Background(), nil, core.CreateWorkbenchInput{
		Namespace:        "ns1",
		WorkbenchName:    "my-wb",
		ImageDisplayName: "PyTorch",
		ImageTag:         "v1",
		PVCName:          "existing-pvc",
	})
	if err != nil {
		t.Fatalf("CreateWorkbench returned error: %v", err)
	}
	if out.Message != "Workbench was succesfully created!" {
		t.Errorf("unexpected message: %q", out.Message)
	}
}

func TestCreateWorkbench_AutoCreatesPVC(t *testing.T) {
	mockWorkbenchHelpers(t)

	setupFakeClient(t)

	_, out, err := CreateWorkbench(context.Background(), nil, core.CreateWorkbenchInput{
		Namespace:        "ns1",
		WorkbenchName:    "auto-pvc-wb",
		ImageDisplayName: "PyTorch",
		ImageTag:         "v1",
	})
	if err != nil {
		t.Fatalf("CreateWorkbench returned error: %v", err)
	}
	if out.Message != "Workbench was succesfully created!" {
		t.Errorf("unexpected message: %q", out.Message)
	}
}

func TestUpdateWorkbench_Image(t *testing.T) {
	mockWorkbenchHelpers(t)

	wb := newDetailedWorkbench("update-wb", "ns1", "alice", "OldImage", "v1", "small", "pvc1", true)
	setupFakeClient(t, wb)

	_, out, err := UpdateWorkbench(context.Background(), nil, core.UpdateWorkbenchInput{
		Namespace:        "ns1",
		WorkbenchName:    "update-wb",
		ImageDisplayName: "NewImage",
		ImageTag:         "v2",
	})
	if err != nil {
		t.Fatalf("UpdateWorkbench returned error: %v", err)
	}
	if out.Message != "Workbench was successfully updated!" {
		t.Errorf("unexpected message: %q", out.Message)
	}
}

func TestUpdateWorkbench_ImageTagOnly(t *testing.T) {
	mockWorkbenchHelpers(t)

	wb := newDetailedWorkbench("update-wb", "ns1", "alice", "OldImage", "v1", "small", "pvc1", true)
	setupFakeClient(t, wb)

	_, out, err := UpdateWorkbench(context.Background(), nil, core.UpdateWorkbenchInput{
		Namespace:     "ns1",
		WorkbenchName: "update-wb",
		ImageTag:      "v2",
	})
	if err != nil {
		t.Fatalf("UpdateWorkbench with imageTag only returned error: %v", err)
	}
	if out.Message != "Workbench was successfully updated!" {
		t.Errorf("unexpected message: %q", out.Message)
	}
}

func TestUpdateWorkbench_NotFound(t *testing.T) {
	mockWorkbenchHelpers(t)

	setupFakeClient(t)

	_, _, err := UpdateWorkbench(context.Background(), nil, core.UpdateWorkbenchInput{
		Namespace:        "ns1",
		WorkbenchName:    "nonexistent",
		ImageDisplayName: "Img",
		ImageTag:         "v1",
	})
	if err == nil {
		t.Fatal("expected error updating nonexistent workbench, got nil")
	}
}
