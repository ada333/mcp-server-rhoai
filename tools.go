package main

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

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

	gvr := schema.GroupVersionResource{Group: "kubeflow.org", Version: "v1", Resource: "notebooks"}
	notebooks, err := dyn.Resource(gvr).Namespace(input.Namespace).List(ctx, metav1.ListOptions{})
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
	clientset, err := LogIntoClusterClientSet()
	if err != nil {
		return nil, WorkbenchesOutput{}, err
	}

	namespaces, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, WorkbenchesOutput{}, fmt.Errorf("failed to list namespaces: %v", err)
	}

	workbenches := ""
	for _, ns := range namespaces.Items {
		namespaceName := ns.GetName()
		_, wbOut, err := ListWorkbenches(ctx, req, WorkBenchesInput{Namespace: namespaceName})
		if err != nil {
			workbenches += fmt.Sprintf("%s: error: %v\n", namespaceName, err)
			continue
		}
		if wbOut.Workbenches == "" {
			workbenches += fmt.Sprintf("%s: none\n", namespaceName)
			continue
		}
		workbenches += fmt.Sprintf("%s:\n%s", namespaceName, wbOut.Workbenches)
	}

	return nil, WorkbenchesOutput{Workbenches: workbenches}, nil
}

func EnableWorkbench(ctx context.Context, req *mcp.CallToolRequest, input WorkBenchesInput) (*mcp.CallToolResult, string, error) {
	return nil, "Workbench enabled", nil
}
