package tools

import (
	"testing"

	core "github.com/amaly/mcp-server-rhoai/core"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

var allListKinds = map[schema.GroupVersionResource]string{
	core.WorkbenchesGVR:     "NotebookList",
	core.PvcGVR:             "PersistentVolumeClaimList",
	core.ImagesGVR:          "ImageStreamList",
	core.HardwareProfilesGVR: "HardwareProfileList",
}

func setupFakeClient(t *testing.T, objects ...runtime.Object) dynamic.Interface {
	orig := GetDynamicClient
	t.Cleanup(func() { GetDynamicClient = orig })
	scheme := runtime.NewScheme()
	client := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(scheme, allListKinds, objects...)
	GetDynamicClient = func() (dynamic.Interface, error) {
		return client, nil
	}
	return client
}

func setupFakeClientSet(t *testing.T, objects ...runtime.Object) kubernetes.Interface {
	orig := GetClientSet
	t.Cleanup(func() { GetClientSet = orig })
	client := fake.NewSimpleClientset(objects...)
	GetClientSet = func() (kubernetes.Interface, error) {
		return client, nil
	}
	return client
}
