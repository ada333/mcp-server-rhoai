package tools

import (
	"context"
	"testing"

	core "github.com/amaly/mcp-server-rhoai/core"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	dynamicfake "k8s.io/client-go/dynamic/fake"
)

func newUnstructuredHardwareProfile(name string, identifiers []interface{}) *unstructured.Unstructured {
	u := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "infrastructure.opendatahub.io/v1",
			"kind":       "HardwareProfile",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": core.GetDefaultNamespace(),
			},
			"spec": map[string]interface{}{
				"identifiers": identifiers,
			},
		},
	}
	return u
}

func makeIdentifier(displayName, identifier, resourceType, defaultCount, maxCount, minCount string) map[string]interface{} {
	return map[string]interface{}{
		"displayName":  displayName,
		"identifier":   identifier,
		"resourceType": resourceType,
		"defaultCount": defaultCount,
		"maxCount":     maxCount,
		"minCount":     minCount,
	}
}

func TestListHardwareProfiles_Success(t *testing.T) {
	orig := GetDynamicClient
	defer func() { GetDynamicClient = orig }()

	scheme := runtime.NewScheme()
	hp := newUnstructuredHardwareProfile("small", []interface{}{
		makeIdentifier("CPU", "cpu", "CPU", "1", "4", "1"),
		makeIdentifier("Memory", "memory", "Memory", "2Gi", "8Gi", "1Gi"),
	})

	client := dynamicfake.NewSimpleDynamicClient(scheme, hp)
	GetDynamicClient = func() (dynamic.Interface, error) {
		return client, nil
	}

	_, out, err := ListHardwareProfiles(context.Background(), nil, struct{}{})
	if err != nil {
		t.Fatalf("ListHardwareProfiles returned error: %v", err)
	}
	if len(out.HardwareProfiles) != 1 {
		t.Fatalf("expected 1 hardware profile, got %d", len(out.HardwareProfiles))
	}
	if out.HardwareProfiles[0].HardwareProfileName != "small" {
		t.Errorf("expected name 'small', got %q", out.HardwareProfiles[0].HardwareProfileName)
	}
	if len(out.HardwareProfiles[0].Resources) != 2 {
		t.Errorf("expected 2 resources, got %d", len(out.HardwareProfiles[0].Resources))
	}
}

func TestListHardwareProfiles_Empty(t *testing.T) {
	orig := GetDynamicClient
	defer func() { GetDynamicClient = orig }()

	scheme := runtime.NewScheme()
	client := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(scheme,
		map[schema.GroupVersionResource]string{
			core.HardwareProfilesGVR: "HardwareProfileList",
		},
	)
	GetDynamicClient = func() (dynamic.Interface, error) {
		return client, nil
	}

	_, out, err := ListHardwareProfiles(context.Background(), nil, struct{}{})
	if err != nil {
		t.Fatalf("ListHardwareProfiles returned error: %v", err)
	}
	if len(out.HardwareProfiles) != 0 {
		t.Errorf("expected 0 hardware profiles, got %d", len(out.HardwareProfiles))
	}
}

func TestCreateHardwareProfile_Success(t *testing.T) {
	orig := GetDynamicClient
	defer func() { GetDynamicClient = orig }()

	scheme := runtime.NewScheme()
	client := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(scheme,
		map[schema.GroupVersionResource]string{
			core.HardwareProfilesGVR: "HardwareProfileList",
		},
	)
	GetDynamicClient = func() (dynamic.Interface, error) {
		return client, nil
	}

	input := core.HardwareProfile{
		HardwareProfileName: "medium",
		Resources: []core.HardwareProfileResource{
			{ResourceName: "CPU", ResourceIdentifier: "cpu", ResourceType: "CPU", DefaultCount: "2", MaxCount: "8", MinCount: "1"},
		},
	}

	_, out, err := CreateHardwareProfile(context.Background(), nil, input)
	if err != nil {
		t.Fatalf("CreateHardwareProfile returned error: %v", err)
	}
	if out.Message != "Hardware Profile was successfully created!" {
		t.Errorf("unexpected message: %q", out.Message)
	}
}

func TestDeleteHardwareProfile_Success(t *testing.T) {
	orig := GetDynamicClient
	defer func() { GetDynamicClient = orig }()

	scheme := runtime.NewScheme()
	hp := newUnstructuredHardwareProfile("to-delete", []interface{}{
		makeIdentifier("CPU", "cpu", "CPU", "1", "2", "1"),
	})
	client := dynamicfake.NewSimpleDynamicClient(scheme, hp)
	GetDynamicClient = func() (dynamic.Interface, error) {
		return client, nil
	}

	_, out, err := DeleteHardwareProfile(context.Background(), nil, core.DeleteHardwareProfileInput{HardwareProfileName: "to-delete"})
	if err != nil {
		t.Fatalf("DeleteHardwareProfile returned error: %v", err)
	}
	if out.Message != "Hardware Profile to-delete was successfully deleted" {
		t.Errorf("unexpected message: %q", out.Message)
	}
}

func TestDeleteHardwareProfile_NotFound(t *testing.T) {
	orig := GetDynamicClient
	defer func() { GetDynamicClient = orig }()

	scheme := runtime.NewScheme()
	client := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(scheme,
		map[schema.GroupVersionResource]string{
			core.HardwareProfilesGVR: "HardwareProfileList",
		},
	)
	GetDynamicClient = func() (dynamic.Interface, error) {
		return client, nil
	}

	_, _, err := DeleteHardwareProfile(context.Background(), nil, core.DeleteHardwareProfileInput{HardwareProfileName: "nonexistent"})
	if err == nil {
		t.Fatal("expected error deleting nonexistent hardware profile, got nil")
	}
}

func TestUpdateHardwareProfile_Success(t *testing.T) {
	orig := GetDynamicClient
	defer func() { GetDynamicClient = orig }()

	scheme := runtime.NewScheme()
	hp := newUnstructuredHardwareProfile("updatable", []interface{}{
		makeIdentifier("CPU", "cpu", "CPU", "1", "4", "1"),
	})
	client := dynamicfake.NewSimpleDynamicClient(scheme, hp)
	GetDynamicClient = func() (dynamic.Interface, error) {
		return client, nil
	}

	input := core.UpdateHardwareProfileInput{
		HardwareProfileName:    "updatable",
		NewHardwareProfileName: "updated-name",
		Resources: []core.HardwareProfileResource{
			{ResourceName: "Memory", ResourceIdentifier: "memory", ResourceType: "Memory", DefaultCount: "4Gi", MaxCount: "16Gi", MinCount: "2Gi"},
		},
	}

	_, out, err := UpdateHardwareProfile(context.Background(), nil, input)
	if err != nil {
		t.Fatalf("UpdateHardwareProfile returned error: %v", err)
	}
	if out.Message != "Hardware Profile was successfully updated!" {
		t.Errorf("unexpected message: %q", out.Message)
	}
}

func TestGetResourcesFromHardwareProfile(t *testing.T) {
	hp := newUnstructuredHardwareProfile("test", []interface{}{
		makeIdentifier("CPU", "cpu", "CPU", "2", "8", "1"),
		makeIdentifier("Memory", "memory", "Memory", "4Gi", "16Gi", "2Gi"),
	})

	resources, err := GetResourcesFromHardwareProfile(hp)
	if err != nil {
		t.Fatalf("GetResourcesFromHardwareProfile returned error: %v", err)
	}
	if len(resources) != 2 {
		t.Fatalf("expected 2 resources, got %d", len(resources))
	}

	cpuFound := false
	memFound := false
	for _, r := range resources {
		if r.ResourceIdentifier == "cpu" {
			cpuFound = true
			if r.DefaultCount != "2" {
				t.Errorf("expected CPU default 2, got %q", r.DefaultCount)
			}
		}
		if r.ResourceIdentifier == "memory" {
			memFound = true
			if r.DefaultCount != "4Gi" {
				t.Errorf("expected memory default 4Gi, got %q", r.DefaultCount)
			}
		}
	}
	if !cpuFound {
		t.Error("cpu resource not found")
	}
	if !memFound {
		t.Error("memory resource not found")
	}
}

func TestGetResourcesFromHardwareProfile_NoIdentifiers(t *testing.T) {
	hp := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "infrastructure.opendatahub.io/v1",
			"kind":       "HardwareProfile",
			"metadata": map[string]interface{}{
				"name":      "empty",
				"namespace": core.GetDefaultNamespace(),
			},
			"spec": map[string]interface{}{},
		},
	}

	_, err := GetResourcesFromHardwareProfile(hp)
	if err == nil {
		t.Error("expected error for hardware profile with no identifiers, got nil")
	}
}
