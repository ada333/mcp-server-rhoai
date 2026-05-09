package tools

import (
	"context"
	"strings"
	"testing"

	core "github.com/amaly/mcp-server-rhoai/core"
	"github.com/amaly/mcp-server-rhoai/resources"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	dynamicfake "k8s.io/client-go/dynamic/fake"
)

func newUnstructuredImageForToolTest(name, displayName, repoURL string, versions []string) *unstructured.Unstructured {
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(core.ImagesGVR.GroupVersion().WithKind("ImageStream"))
	u.SetName(name)
	u.SetNamespace(core.GetDefaultNamespace())
	u.SetLabels(map[string]string{
		"opendatahub.io/notebook-image": "true",
	})
	u.SetAnnotations(map[string]string{
		"opendatahub.io/notebook-image-name": displayName,
	})

	if repoURL != "" {
		if err := unstructured.SetNestedField(u.Object, repoURL, "status", "dockerImageRepository"); err != nil {
			panic(err)
		}
	}

	tags := make([]interface{}, len(versions))
	for i, v := range versions {
		tags[i] = map[string]interface{}{"name": v}
	}
	if err := unstructured.SetNestedSlice(u.Object, tags, "spec", "tags"); err != nil {
		panic(err)
	}

	return u
}

func newUnstructuredImage(name string, isDefault bool, annotations map[string]string) *unstructured.Unstructured {
	u := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "image.openshift.io/v1",
			"kind":       "ImageStream",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": core.GetDefaultNamespace(),
			},
		},
	}

	labels := map[string]string{
		"opendatahub.io/notebook-image": "true",
	}
	u.SetLabels(labels)

	if annotations == nil {
		annotations = map[string]string{}
	}
	if isDefault {
		annotations["internal.config.kubernetes.io/previousNamespaces"] = "default"
	}
	u.SetAnnotations(annotations)

	return u
}

func newWorkbenchWithImageAnnotation(name, namespace, imageName string) *unstructured.Unstructured {
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(core.WorkbenchesGVR.GroupVersion().WithKind("Notebook"))
	u.SetName(name)
	u.SetNamespace(namespace)
	u.SetAnnotations(map[string]string{
		"opendatahub.io/image-display-name": imageName,
	})
	return u
}

func TestCreateCustomImage_Success(t *testing.T) {
	setupFakeClient(t)

	input := core.CreateCustomImageInput{
		ImageName:        "my-custom-image",
		ImageLocation:    "quay.io/my/image:latest",
		ImageDescription: "A custom test image",
	}

	_, out, err := CreateCustomImage(context.Background(), nil, input)
	if err != nil {
		t.Fatalf("CreateCustomImage returned error: %v", err)
	}
	if out.Message != "Image was successfully created!" {
		t.Errorf("unexpected message: %q", out.Message)
	}
}

func TestUpdateImage_Success(t *testing.T) {
	img := newUnstructuredImage("img-to-update", false, map[string]string{
		"opendatahub.io/notebook-image-name": "OldName",
	})
	setupFakeClient(t, img)

	_, out, err := UpdateImage(context.Background(), nil, core.UpdateImageInput{
		ImageName:    "img-to-update",
		NewImageName: "NewName",
	})
	if err != nil {
		t.Fatalf("UpdateImage returned error: %v", err)
	}
	if !strings.Contains(out.Message, "successfully updated") {
		t.Errorf("unexpected message: %q", out.Message)
	}
}

func TestUpdateImage_NotFound(t *testing.T) {
	setupFakeClient(t)

	_, _, err := UpdateImage(context.Background(), nil, core.UpdateImageInput{
		ImageName:    "nonexistent",
		NewImageName: "new-name",
	})
	if err == nil {
		t.Fatal("expected error updating nonexistent image, got nil")
	}
}

func TestDeleteImage_Success(t *testing.T) {
	img := newUnstructuredImage("custom-img", false, map[string]string{
		"opendatahub.io/notebook-image-name": "CustomImage",
	})
	setupFakeClient(t, img)

	_, out, err := DeleteImage(context.Background(), nil, core.DeleteImageInput{ImageName: "custom-img"})
	if err != nil {
		t.Fatalf("DeleteImage returned error: %v", err)
	}
	if !strings.Contains(out.Message, "successfully deleted") {
		t.Errorf("unexpected message: %q", out.Message)
	}
}

func TestDeleteImage_DefaultImageBlocked(t *testing.T) {
	img := newUnstructuredImage("default-img", true, map[string]string{
		"opendatahub.io/notebook-image-name": "DefaultImage",
	})
	setupFakeClient(t, img)

	_, _, err := DeleteImage(context.Background(), nil, core.DeleteImageInput{ImageName: "default-img"})
	if err == nil {
		t.Fatal("expected error deleting default image, got nil")
	}
	if !strings.Contains(err.Error(), "default image") {
		t.Errorf("expected 'default image' in error, got: %v", err)
	}
}

func TestDeleteImage_UsedImageBlocked(t *testing.T) {
	img := newUnstructuredImage("used-img", false, map[string]string{
		"opendatahub.io/notebook-image-name": "UsedImage",
	})
	wb := newWorkbenchWithImageAnnotation("wb1", "ns1", "used-img")
	setupFakeClient(t, img, wb)

	_, _, err := DeleteImage(context.Background(), nil, core.DeleteImageInput{ImageName: "used-img"})
	if err == nil {
		t.Fatal("expected error deleting used image, got nil")
	}
	if !strings.Contains(err.Error(), "used by a workbench") {
		t.Errorf("expected 'used by a workbench' in error, got: %v", err)
	}
}

func TestImageIsDefault_True(t *testing.T) {
	img := newUnstructuredImage("default-img", true, nil)
	setupFakeClient(t, img)

	result, err := ImageIsDefault(context.Background(), "default-img")
	if err != nil {
		t.Fatalf("ImageIsDefault returned error: %v", err)
	}
	if !result {
		t.Error("expected ImageIsDefault to return true for default image")
	}
}

func TestImageIsDefault_False(t *testing.T) {
	img := newUnstructuredImage("custom-img", false, nil)
	setupFakeClient(t, img)

	result, err := ImageIsDefault(context.Background(), "custom-img")
	if err != nil {
		t.Fatalf("ImageIsDefault returned error: %v", err)
	}
	if result {
		t.Error("expected ImageIsDefault to return false for custom image")
	}
}

func TestImageIsUsed_True(t *testing.T) {
	wb := newWorkbenchWithImageAnnotation("wb1", "ns1", "my-image")
	setupFakeClient(t, wb)

	result, err := ImageIsUsed(context.Background(), "my-image")
	if err != nil {
		t.Fatalf("ImageIsUsed returned error: %v", err)
	}
	if !result {
		t.Error("expected ImageIsUsed to return true")
	}
}

func TestImageIsUsed_False(t *testing.T) {
	wb := newWorkbenchWithImageAnnotation("wb1", "ns1", "other-image")
	setupFakeClient(t, wb)

	result, err := ImageIsUsed(context.Background(), "unused-image")
	if err != nil {
		t.Fatalf("ImageIsUsed returned error: %v", err)
	}
	if result {
		t.Error("expected ImageIsUsed to return false")
	}
}

func TestListImages(t *testing.T) {
	orig := resources.GetDynamicClient
	defer func() { resources.GetDynamicClient = orig }()

	scheme := runtime.NewScheme()

	image1 := newUnstructuredImageForToolTest("img1", "PyTorch", "quay.io/modh/pytorch", []string{"v1", "v2"})
	image2 := newUnstructuredImageForToolTest("img2", "TensorFlow", "quay.io/modh/tensorflow", []string{"latest"})

	client := dynamicfake.NewSimpleDynamicClient(scheme, image1, image2)

	resources.GetDynamicClient = func() (dynamic.Interface, error) {
		return client, nil
	}

	_, out, err := ListImages(context.Background(), nil, struct{}{})
	if err != nil {
		t.Fatalf("ListImages returned error: %v", err)
	}

	if len(out.Images) != 2 {
		t.Fatalf("expected 2 images, got %d", len(out.Images))
	}

	// Find PyTorch image
	var pytorch *core.ImageInfo
	var tensorflow *core.ImageInfo
	for i := range out.Images {
		if out.Images[i].Name == "PyTorch" {
			pytorch = &out.Images[i]
		}
		if out.Images[i].Name == "TensorFlow" {
			tensorflow = &out.Images[i]
		}
	}

	if pytorch == nil {
		t.Errorf("expected PyTorch image in output")
	} else {
		if pytorch.URL != "quay.io/modh/pytorch" {
			t.Errorf("expected PyTorch URL quay.io/modh/pytorch, got: %s", pytorch.URL)
		}
		if len(pytorch.Versions) != 2 {
			t.Errorf("expected 2 versions for PyTorch, got %d", len(pytorch.Versions))
		}
	}

	if tensorflow == nil {
		t.Errorf("expected TensorFlow image in output")
	}
}
