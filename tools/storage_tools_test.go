package tools

import (
	"context"
	"strings"
	"testing"

	core "github.com/amaly/mcp-server-rhoai/core"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	dynamicfake "k8s.io/client-go/dynamic/fake"
)

func newUnstructuredPVC(name, namespace, size string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "PersistentVolumeClaim",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
			},
			"spec": map[string]interface{}{
				"accessModes": []interface{}{"ReadWriteOnce"},
				"resources": map[string]interface{}{
					"requests": map[string]interface{}{
						"storage": size,
					},
				},
			},
		},
	}
}

func TestListPVCs_Success(t *testing.T) {
	orig := GetDynamicClient
	defer func() { GetDynamicClient = orig }()

	scheme := runtime.NewScheme()
	client := dynamicfake.NewSimpleDynamicClient(scheme,
		newUnstructuredPVC("pvc-a", "ns1", "10Gi"),
		newUnstructuredPVC("pvc-b", "ns1", "20Gi"),
		newUnstructuredPVC("pvc-other", "ns2", "5Gi"),
	)
	GetDynamicClient = func() (dynamic.Interface, error) {
		return client, nil
	}

	_, out, err := ListPVCs(context.Background(), nil, core.ListPVCsInput{Namespace: "ns1"})
	if err != nil {
		t.Fatalf("ListPVCs returned error: %v", err)
	}
	if !strings.Contains(out.PVCs, "pvc-a") {
		t.Errorf("expected pvc-a in output, got: %q", out.PVCs)
	}
	if !strings.Contains(out.PVCs, "pvc-b") {
		t.Errorf("expected pvc-b in output, got: %q", out.PVCs)
	}
	if strings.Contains(out.PVCs, "pvc-other") {
		t.Errorf("did not expect pvc-other (different namespace) in output, got: %q", out.PVCs)
	}
}

func TestListPVCs_Empty(t *testing.T) {
	orig := GetDynamicClient
	defer func() { GetDynamicClient = orig }()

	scheme := runtime.NewScheme()
	client := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(scheme,
		map[schema.GroupVersionResource]string{
			core.PvcGVR: "PersistentVolumeClaimList",
		},
	)
	GetDynamicClient = func() (dynamic.Interface, error) {
		return client, nil
	}

	_, out, err := ListPVCs(context.Background(), nil, core.ListPVCsInput{Namespace: "empty-ns"})
	if err != nil {
		t.Fatalf("ListPVCs returned error: %v", err)
	}
	if out.PVCs != "" {
		t.Errorf("expected empty output, got: %q", out.PVCs)
	}
}

func TestCreatePVC_Success(t *testing.T) {
	orig := GetDynamicClient
	defer func() { GetDynamicClient = orig }()

	scheme := runtime.NewScheme()
	client := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(scheme,
		map[schema.GroupVersionResource]string{
			core.PvcGVR: "PersistentVolumeClaimList",
		},
	)
	GetDynamicClient = func() (dynamic.Interface, error) {
		return client, nil
	}

	_, out, err := CreatePVC(context.Background(), nil, core.PVCInput{
		Namespace: "ns1",
		PVCName:   "new-pvc",
		Size:      "10Gi",
	})
	if err != nil {
		t.Fatalf("CreatePVC returned error: %v", err)
	}
	if out.Message != "PVC was succesfully created!" {
		t.Errorf("unexpected message: %q", out.Message)
	}
}

func TestCreatePVC_Duplicate(t *testing.T) {
	orig := GetDynamicClient
	defer func() { GetDynamicClient = orig }()

	scheme := runtime.NewScheme()
	client := dynamicfake.NewSimpleDynamicClient(scheme, newUnstructuredPVC("existing", "ns1", "10Gi"))
	GetDynamicClient = func() (dynamic.Interface, error) {
		return client, nil
	}

	_, _, err := CreatePVC(context.Background(), nil, core.PVCInput{
		Namespace: "ns1",
		PVCName:   "existing",
		Size:      "10Gi",
	})
	if err == nil {
		t.Fatal("expected error creating duplicate PVC, got nil")
	}
}

func TestDeletePVC_Success(t *testing.T) {
	orig := GetDynamicClient
	defer func() { GetDynamicClient = orig }()

	scheme := runtime.NewScheme()
	client := dynamicfake.NewSimpleDynamicClient(scheme, newUnstructuredPVC("to-delete", "ns1", "10Gi"))
	GetDynamicClient = func() (dynamic.Interface, error) {
		return client, nil
	}

	_, out, err := DeletePVC(context.Background(), nil, core.DeletePVCInput{Namespace: "ns1", PVCName: "to-delete"})
	if err != nil {
		t.Fatalf("DeletePVC returned error: %v", err)
	}
	if out.Message != "PVC was successfully deleted!" {
		t.Errorf("unexpected message: %q", out.Message)
	}
}

func TestDeletePVC_NotFound(t *testing.T) {
	orig := GetDynamicClient
	defer func() { GetDynamicClient = orig }()

	scheme := runtime.NewScheme()
	client := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(scheme,
		map[schema.GroupVersionResource]string{
			core.PvcGVR: "PersistentVolumeClaimList",
		},
	)
	GetDynamicClient = func() (dynamic.Interface, error) {
		return client, nil
	}

	_, _, err := DeletePVC(context.Background(), nil, core.DeletePVCInput{Namespace: "ns1", PVCName: "nonexistent"})
	if err == nil {
		t.Fatal("expected error deleting nonexistent PVC, got nil")
	}
}

func TestUpdatePVC_IncreaseSize(t *testing.T) {
	orig := GetDynamicClient
	defer func() { GetDynamicClient = orig }()

	scheme := runtime.NewScheme()
	client := dynamicfake.NewSimpleDynamicClient(scheme, newUnstructuredPVC("resize-me", "ns1", "10Gi"))
	GetDynamicClient = func() (dynamic.Interface, error) {
		return client, nil
	}

	_, out, err := UpdatePVC(context.Background(), nil, core.PVCInput{
		Namespace: "ns1",
		PVCName:   "resize-me",
		Size:      "20Gi",
	})
	if err != nil {
		t.Fatalf("UpdatePVC returned error: %v", err)
	}
	if !strings.Contains(out.Message, "successfully updated") {
		t.Errorf("unexpected message: %q", out.Message)
	}
}

func TestUpdatePVC_DecreaseSize(t *testing.T) {
	orig := GetDynamicClient
	defer func() { GetDynamicClient = orig }()

	scheme := runtime.NewScheme()
	client := dynamicfake.NewSimpleDynamicClient(scheme, newUnstructuredPVC("shrink-me", "ns1", "20Gi"))
	GetDynamicClient = func() (dynamic.Interface, error) {
		return client, nil
	}

	_, _, err := UpdatePVC(context.Background(), nil, core.PVCInput{
		Namespace: "ns1",
		PVCName:   "shrink-me",
		Size:      "5Gi",
	})
	if err == nil {
		t.Fatal("expected error when decreasing PVC size, got nil")
	}
	if !strings.Contains(err.Error(), "cannot be decreased") {
		t.Errorf("expected 'cannot be decreased' in error, got: %v", err)
	}
}

func TestUpdatePVC_NotFound(t *testing.T) {
	orig := GetDynamicClient
	defer func() { GetDynamicClient = orig }()

	scheme := runtime.NewScheme()
	client := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(scheme,
		map[schema.GroupVersionResource]string{
			core.PvcGVR: "PersistentVolumeClaimList",
		},
	)
	GetDynamicClient = func() (dynamic.Interface, error) {
		return client, nil
	}

	_, _, err := UpdatePVC(context.Background(), nil, core.PVCInput{
		Namespace: "ns1",
		PVCName:   "nonexistent",
		Size:      "10Gi",
	})
	if err == nil {
		t.Fatal("expected error updating nonexistent PVC, got nil")
	}
}
