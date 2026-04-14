package tools

import (
	"context"
	"fmt"

	"github.com/amaly/mcp-server-rhoai/core"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
)

func ListPVCs(ctx context.Context, req *mcp.CallToolRequest, input core.ListPVCsInput) (*mcp.CallToolResult, core.PVCsOutput, error) {
	dyn, err := GetDynamicClient()
	if err != nil {
		return nil, core.PVCsOutput{}, err
	}

	pvcs, err := dyn.Resource(core.PvcGVR).Namespace(input.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, core.PVCsOutput{}, fmt.Errorf("failed to list PVCs: %v", err)
	}

	msg := ""
	for _, pvc := range pvcs.Items {
		msg += fmt.Sprintf("- %s\n", pvc.GetName())
	}
	return nil, core.PVCsOutput{PVCs: msg}, nil
}

func CreatePVC(ctx context.Context, req *mcp.CallToolRequest, input core.PVCInput) (*mcp.CallToolResult, core.DefaultToolOutput, error) {
	dyn, err := GetDynamicClient()
	if err != nil {
		return nil, core.DefaultToolOutput{}, err
	}

	err = createPersistentVolumeClaim(ctx, dyn, input.Namespace, input.PVCName, input.Size)
	if err != nil {
		return nil, core.DefaultToolOutput{}, fmt.Errorf("failed to create PVC: %v", err)
	}

	return nil, core.DefaultToolOutput{Message: "PVC was succesfully created!"}, nil
}

func UpdatePVC(ctx context.Context, req *mcp.CallToolRequest, input core.PVCInput) (*mcp.CallToolResult, core.DefaultToolOutput, error) {
	dyn, err := GetDynamicClient()
	if err != nil {
		return nil, core.DefaultToolOutput{}, err
	}

	pvc, err := dyn.Resource(core.PvcGVR).Namespace(input.Namespace).Get(ctx, input.PVCName, metav1.GetOptions{})
	if err != nil {
		return nil, core.DefaultToolOutput{}, fmt.Errorf("failed to get PVC: %v", err)
	}

	current_size, err := getDiskUsageFromPVC(ctx, dyn, input.Namespace, input.PVCName)
	if err != nil {
		return nil, core.DefaultToolOutput{}, fmt.Errorf("failed to get PVC size: %v", err)
	}
	if input.Size != "" {
		newQty, err := resource.ParseQuantity(input.Size)
		if err != nil {
			return nil, core.DefaultToolOutput{}, fmt.Errorf("invalid PVC size %q: %v", input.Size, err)
		}
		currentQty, err := resource.ParseQuantity(current_size)
		if err != nil {
			return nil, core.DefaultToolOutput{}, fmt.Errorf("invalid current PVC size %q: %v", current_size, err)
		}
		if newQty.Cmp(currentQty) < 0 {
			return nil, core.DefaultToolOutput{}, fmt.Errorf("PVC size cannot be decreased - current size is %s and new size is %s", current_size, input.Size)
		}
	}
	if input.Size != "" {
		if err := unstructured.SetNestedField(pvc.Object, input.Size, "spec", "resources", "requests", "storage"); err != nil {
			return nil, core.DefaultToolOutput{}, fmt.Errorf("failed to set PVC size: %v", err)
		}
	}
	if input.NewPVCName != "" {
		pvc.SetAnnotations(map[string]string{
			"openshift.io/display-name": input.NewPVCName,
		})
	}
	_, err = dyn.Resource(core.PvcGVR).Namespace(input.Namespace).Update(ctx, pvc, metav1.UpdateOptions{})
	if err != nil {
		return nil, core.DefaultToolOutput{}, fmt.Errorf("failed to update PVC: %v", err)
	}

	return nil, core.DefaultToolOutput{Message: "PVC was successfully updated - please note that if you increased the PVC size, the workbench will restart and be unavailable for a period of time that is usually proportional to the size change."}, nil
}

func createPersistentVolumeClaim(ctx context.Context, dyn dynamic.Interface, namespace, name, size string) error {
	pvc := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "PersistentVolumeClaim",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
				"labels": map[string]interface{}{
					"opendatahub.io/dashboard": "true",
				},
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

	_, err := dyn.Resource(core.PvcGVR).Namespace(namespace).Create(ctx, pvc, metav1.CreateOptions{})
	return err
}

func DeletePVC(ctx context.Context, req *mcp.CallToolRequest, input core.DeletePVCInput) (*mcp.CallToolResult, core.DefaultToolOutput, error) {
	dyn, err := GetDynamicClient()
	if err != nil {
		return nil, core.DefaultToolOutput{}, err
	}

	err = dyn.Resource(core.PvcGVR).Namespace(input.Namespace).Delete(ctx, input.PVCName, metav1.DeleteOptions{})
	if err != nil {
		return nil, core.DefaultToolOutput{}, fmt.Errorf("failed to delete PVC: %v", err)
	}

	return nil, core.DefaultToolOutput{Message: "PVC was successfully deleted!"}, nil
}

func getDiskUsageFromPVC(ctx context.Context, dyn dynamic.Interface, namespace, pvcName string) (string, error) {
	pvc, err := dyn.Resource(core.PvcGVR).Namespace(namespace).Get(ctx, pvcName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get PVC: %v", err)
	}
	capacity, found, err := unstructured.NestedString(pvc.Object, "spec", "resources", "requests", "storage")
	if err != nil {
		return "", fmt.Errorf("failed to get storage capacity: %v", err)
	}
	if !found {
		return "", nil
	}
	return capacity, nil
}
