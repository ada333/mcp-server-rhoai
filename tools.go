package main

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8stypes "k8s.io/apimachinery/pkg/types"
)

var workbenchesGVR = schema.GroupVersionResource{Group: "kubeflow.org", Version: "v1", Resource: "notebooks"}

func ListPods(ctx context.Context, req *mcp.CallToolRequest, input WorkBenchesInput) (*mcp.CallToolResult, PodsOutput, error) {
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

func ListWorkbenches(ctx context.Context, req *mcp.CallToolRequest, input WorkBenchesInput) (*mcp.CallToolResult, WorkbenchesOutput, error) {

	dyn, err := LogIntoClusterDynamic()
	if err != nil {
		return nil, WorkbenchesOutput{}, err
	}

	notebooks, err := dyn.Resource(workbenchesGVR).Namespace(input.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, WorkbenchesOutput{}, fmt.Errorf("failed to list workbenches: %v", err)
	}

	msg := ""
	for _, nb := range notebooks.Items {
		name := nb.GetName()
		msg += fmt.Sprintf("- %s\n", name)
	}
	return nil, WorkbenchesOutput{Workbenches: msg}, nil
}

func ListAllWorkbenches(ctx context.Context, req *mcp.CallToolRequest, input WorkBenchesInput) (*mcp.CallToolResult, WorkbenchesOutput, error) {
	_, workbenches, err := ListWorkbenches(ctx, req, WorkBenchesInput{Namespace: ""})
	if err != nil {
		return nil, WorkbenchesOutput{}, err
	}
	return nil, WorkbenchesOutput{Workbenches: workbenches.Workbenches}, nil
}

func EnableWorkbench(ctx context.Context, req *mcp.CallToolRequest, input WorkbenchToggleInput) (*mcp.CallToolResult, string, error) {
	dyn, err := LogIntoClusterDynamic()
	if err != nil {
		return nil, "", err
	}

	state := "running"
	action := "enabled"
	if input.Disable {
		state = "stopped"
		action = "disabled"
	}

	patch := fmt.Sprintf(`{"metadata":{"annotations":{"notebooks.kubeflow.org/notebook-state":"%s"}}}`, state)

	_, err = dyn.Resource(workbenchesGVR).Namespace(input.Namespace).Patch(
		ctx,
		input.WorkbenchName,
		k8stypes.MergePatchType,
		[]byte(patch),
		metav1.PatchOptions{},
	)
	if err != nil {
		return nil, "", fmt.Errorf("failed to %s workbench %q: %v", action, input.WorkbenchName, err)
	}

	return nil, fmt.Sprintf("Workbench %s %s", input.WorkbenchName, action), nil
}
