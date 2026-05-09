package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	core "github.com/amaly/mcp-server-rhoai/core"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8stypes "k8s.io/apimachinery/pkg/types"
)

func ListWorkbenches(ctx context.Context, req *mcp.CallToolRequest, input core.ListWorkbenchesInput) (*mcp.CallToolResult, core.ListWorkbenchesResult, error) {
	dyn, err := GetDynamicClient()
	if err != nil {
		return nil, core.ListWorkbenchesResult{}, err
	}

	workbenches, err := dyn.Resource(core.WorkbenchesGVR).Namespace(input.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, core.ListWorkbenchesResult{}, fmt.Errorf("failed to list workbenches: %v", err)
	}

	workbenchesInfo := make([]core.WorkbenchInfo, 0, len(workbenches.Items))
	for _, wb := range workbenches.Items {
		info, err := extractWorkbenchInfo(ctx, dyn, wb)
		if err != nil {
			return nil, core.ListWorkbenchesResult{}, err
		}
		workbenchesInfo = append(workbenchesInfo, info)
	}
	return nil, core.ListWorkbenchesResult{Workbenches: workbenchesInfo}, nil
}

func CreateWorkbench(ctx context.Context, req *mcp.CallToolRequest, input core.CreateWorkbenchInput) (*mcp.CallToolResult, core.DefaultToolOutput, error) {
	dyn, err := GetDynamicClient()
	if err != nil {
		return nil, core.DefaultToolOutput{}, err
	}

	repoURL, gitCommit, imageName, err := getImageInfoFn(ctx, input.ImageDisplayName, input.ImageTag)
	if err != nil {
		return nil, core.DefaultToolOutput{}, fmt.Errorf("failed to lookup image info: %v", err)
	}

	if input.PVCName == "" {
		input.PVCName = input.WorkbenchName
		if err := createPersistentVolumeClaim(ctx, dyn, input.Namespace, input.PVCName, defaultPVCSize); err != nil {
			return nil, core.DefaultToolOutput{}, fmt.Errorf("failed to create PVC: %v, try using a different workbench name", err)
		}
	}

	imageFull := resolveFullImageURL(repoURL, input.ImageTag)
	hardwareProfile := resolveHardwareProfile(input.HardwareProfile)
	limits, requests := buildResourceRequirements(hardwareProfile)
	notebook := buildNotebookObject(input, imageFull, imageName, gitCommit, hardwareProfile, limits, requests)

	_, err = dyn.Resource(core.WorkbenchesGVR).Namespace(input.Namespace).Create(ctx, notebook, metav1.CreateOptions{})
	if err != nil {
		return nil, core.DefaultToolOutput{}, fmt.Errorf("failed to create notebook: %v", err)
	}

	return nil, core.DefaultToolOutput{Message: "Workbench was succesfully created!"}, nil
}

func UpdateWorkbench(ctx context.Context, req *mcp.CallToolRequest, input core.UpdateWorkbenchInput) (*mcp.CallToolResult, core.DefaultToolOutput, error) {
	dyn, err := GetDynamicClient()
	if err != nil {
		return nil, core.DefaultToolOutput{}, err
	}

	workbench, err := dyn.Resource(core.WorkbenchesGVR).Namespace(input.Namespace).Get(ctx, input.WorkbenchName, metav1.GetOptions{})
	if err != nil {
		return nil, core.DefaultToolOutput{}, fmt.Errorf("failed to get workbench %s: %v", input.WorkbenchName, err)
	}

	annotations := workbench.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}

	container, containers, err := getFirstContainer(workbench)
	if err != nil {
		return nil, core.DefaultToolOutput{}, fmt.Errorf("workbench %s: %v", input.WorkbenchName, err)
	}

	if input.ImageDisplayName != "" && input.ImageTag != "" {
		if err := updateWorkbenchImage(ctx, container, annotations, input.ImageDisplayName, input.ImageTag); err != nil {
			return nil, core.DefaultToolOutput{}, err
		}
	}

	if input.HardwareProfile.HardwareProfileName != "" {
		updateWorkbenchHardwareProfile(container, annotations, input.HardwareProfile)
	}

	if input.PVCName != "" {
		if err := updateWorkbenchPVC(container, workbench, input.PVCName); err != nil {
			return nil, core.DefaultToolOutput{}, fmt.Errorf("failed to update PVC: %v", err)
		}
	}

	containers[0] = container
	if err := unstructured.SetNestedSlice(workbench.Object, containers, "spec", "template", "spec", "containers"); err != nil {
		return nil, core.DefaultToolOutput{}, fmt.Errorf("failed to set containers: %v", err)
	}

	workbench.SetAnnotations(annotations)

	_, err = dyn.Resource(core.WorkbenchesGVR).Namespace(input.Namespace).Update(ctx, workbench, metav1.UpdateOptions{})
	if err != nil {
		return nil, core.DefaultToolOutput{}, fmt.Errorf("failed to update workbench %s: %v", input.WorkbenchName, err)
	}

	return nil, core.DefaultToolOutput{Message: "Workbench was successfully updated!"}, nil
}

func DeleteWorkbench(ctx context.Context, req *mcp.CallToolRequest, input core.DeleteWorkbenchInput) (*mcp.CallToolResult, core.DefaultToolOutput, error) {
	dyn, err := GetDynamicClient()
	if err != nil {
		return nil, core.DefaultToolOutput{}, err
	}

	err = dyn.Resource(core.WorkbenchesGVR).Namespace(input.Namespace).Delete(ctx, input.WorkbenchName, metav1.DeleteOptions{})
	if err != nil {
		return nil, core.DefaultToolOutput{}, fmt.Errorf("failed to delete workbench %s: %v", input.WorkbenchName, err)
	}

	return nil, core.DefaultToolOutput{Message: fmt.Sprintf("Workbench %s was successfully deleted", input.WorkbenchName)}, nil
}

func ChangeWorkbenchStatus(ctx context.Context, req *mcp.CallToolRequest, input core.ChangeWorkbenchStatusInput) (*mcp.CallToolResult, core.DefaultToolOutput, error) {
	dyn, err := GetDynamicClient()
	if err != nil {
		return nil, core.DefaultToolOutput{}, err
	}

	stopped, err := IsWorkbenchStopped(ctx, dyn, input.Namespace, input.WorkbenchName)
	if err != nil {
		return nil, core.DefaultToolOutput{}, err
	}
	if (input.Status == core.Stopped && stopped) || (input.Status == core.Running && !stopped) {
		return nil, core.DefaultToolOutput{Message: fmt.Sprintf("Workbench %s is already %s", input.WorkbenchName, input.Status)}, nil
	}

	patch := map[string]interface{}{}
	patchAnnotations := map[string]interface{}{}
	if input.Status == core.Stopped {
		patchAnnotations["kubeflow-resource-stopped"] = time.Now().UTC().Format(time.RFC3339)
	} else {
		patchAnnotations["kubeflow-resource-stopped"] = nil
	}
	patch["metadata"] = map[string]interface{}{
		"annotations": patchAnnotations,
	}

	patchBytes, err := json.Marshal(patch)
	if err != nil {
		return nil, core.DefaultToolOutput{}, fmt.Errorf("failed to marshal patch: %v", err)
	}

	_, err = dyn.Resource(core.WorkbenchesGVR).Namespace(input.Namespace).Patch(
		ctx,
		input.WorkbenchName,
		k8stypes.MergePatchType,
		patchBytes,
		metav1.PatchOptions{},
	)
	if err != nil {
		return nil, core.DefaultToolOutput{}, fmt.Errorf("failed to %s workbench %s: %v", input.Status, input.WorkbenchName, err)
	}

	return nil, core.DefaultToolOutput{Message: fmt.Sprintf("Workbench %s is %s", input.WorkbenchName, input.Status)}, nil
}
