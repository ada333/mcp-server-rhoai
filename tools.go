package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8stypes "k8s.io/apimachinery/pkg/types"
)

var workbenchesGVR = schema.GroupVersionResource{Group: "kubeflow.org", Version: "v1", Resource: "notebooks"}

func ListPods(ctx context.Context, req *mcp.CallToolRequest, input ListWorkbenchesInput) (*mcp.CallToolResult, PodsOutput, error) {
	clientset, err := LogIntoClusterClientSet()
	if err != nil {
		return nil, PodsOutput{}, err
	}

	// list pods - this should be only code in the func
	pods, err := clientset.CoreV1().Pods(input.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, PodsOutput{}, fmt.Errorf("failed to list pods: %v", err)
	}

	msg := ""
	for _, pod := range pods.Items {
		msg += fmt.Sprintf("- %s (%s)\n", pod.Name, pod.Status.Phase)
	}
	return nil, PodsOutput{Pods: msg}, nil
}

func ListWorkbenches(ctx context.Context, req *mcp.CallToolRequest, input ListWorkbenchesInput) (*mcp.CallToolResult, ListWorkbenchesResult, error) {

	dyn, err := LogIntoClusterDynamic()
	if err != nil {
		return nil, ListWorkbenchesResult{}, err
	}

	notebooks, err := dyn.Resource(workbenchesGVR).Namespace(input.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, ListWorkbenchesResult{}, fmt.Errorf("failed to list workbenches: %v", err)
	}

	msg := ""
	for _, nb := range notebooks.Items {
		name := nb.GetName()
		msg += fmt.Sprintf("- %s\n", name)
	}
	return nil, ListWorkbenchesResult{Workbenches: msg}, nil
}

func ListAllWorkbenches(ctx context.Context, req *mcp.CallToolRequest, input ListWorkbenchesInput) (*mcp.CallToolResult, ListWorkbenchesResult, error) {
	_, workbenches, err := ListWorkbenches(ctx, req, ListWorkbenchesInput{Namespace: ""})
	if err != nil {
		return nil, ListWorkbenchesResult{}, err
	}
	return nil, ListWorkbenchesResult{Workbenches: workbenches.Workbenches}, nil
}

func ChangeWorkbenchStatus(ctx context.Context, req *mcp.CallToolRequest, input ChangeWorkbenchStatusInput) (*mcp.CallToolResult, ChangeWorkbenchStatusOutput, error) {
	dyn, err := LogIntoClusterDynamic()
	if err != nil {
		return nil, ChangeWorkbenchStatusOutput{}, err
	}

	patchObj := map[string]interface{}{}
	annotations := map[string]interface{}{}
	if input.Status == Stopped {
		annotations["kubeflow-resource-stopped"] = time.Now().UTC().Format(time.RFC3339)
	} else {
		annotations["kubeflow-resource-stopped"] = nil
	}
	patchObj["metadata"] = map[string]interface{}{
		"annotations": annotations,
	}

	patchBytes, err := json.Marshal(patchObj)
	if err != nil {
		return nil, ChangeWorkbenchStatusOutput{}, fmt.Errorf("failed to marshal patch: %v", err)
	}

	_, err = dyn.Resource(workbenchesGVR).Namespace(input.Namespace).Patch(
		ctx,
		input.WorkbenchName,
		k8stypes.MergePatchType,
		patchBytes,
		metav1.PatchOptions{},
	)
	if err != nil {
		return nil, ChangeWorkbenchStatusOutput{}, fmt.Errorf("failed to %s workbench %s: %v", input.Status, input.WorkbenchName, err)
	}

	return nil, ChangeWorkbenchStatusOutput{Message: fmt.Sprintf("Workbench %s is %s", input.WorkbenchName, input.Status)}, nil
}
