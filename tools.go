package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type PodsOutput struct {
	Pods string `json:"pods" jsonschema_description:"the list of pods"`
}

type WorkbenchesOutput struct {
	Workbenches string `json:"workbenches" jsonschema_description:"the list of workbenches"`
}

func LogIntoClusterClientSet() (*kubernetes.Clientset, error) {
	kubeconfigPath := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %v", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to log into cluster: %v", err)
	}
	return clientset, nil
}

func ListPods(ctx context.Context, req *mcp.CallToolRequest, input struct{ Namespace string }) (*mcp.CallToolResult, PodsOutput, error) {
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

func LogIntoClusterDynamic() (*dynamic.DynamicClient, error) {
	kubeconfigPath := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %v", err)
	}
	dyn, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to log into cluster: %v", err)
	}
	return dyn, nil
}

func ListWorkbenches(ctx context.Context, req *mcp.CallToolRequest, input struct{ Namespace string }) (*mcp.CallToolResult, WorkbenchesOutput, error) {

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

func ListAllWorkbenches(ctx context.Context, req *mcp.CallToolRequest, input struct{ Namespace string }) (*mcp.CallToolResult, WorkbenchesOutput, error) {
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
		_, wbOut, err := ListWorkbenches(ctx, req, struct{ Namespace string }{Namespace: namespaceName})
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
