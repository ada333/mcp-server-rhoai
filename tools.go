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
	"k8s.io/client-go/dynamic"
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

func IsWorkbenchStopped(ctx context.Context, dyn *dynamic.DynamicClient, namespace, workbenchName string) (bool, error) {
	current, err := dyn.Resource(workbenchesGVR).Namespace(namespace).Get(ctx, workbenchName, metav1.GetOptions{})
	if err != nil {
		return false, fmt.Errorf("failed to get workbench %s: %v", workbenchName, err)
	}
	currentAnnotations := current.GetAnnotations()
	currentStopped := false
	if currentAnnotations != nil {
		if _, ok := currentAnnotations["kubeflow-resource-stopped"]; ok {
			currentStopped = true
		}
	}
	return currentStopped, nil
}

func ChangeWorkbenchStatus(ctx context.Context, req *mcp.CallToolRequest, input ChangeWorkbenchStatusInput) (*mcp.CallToolResult, WorkbenchOutput, error) {
	dyn, err := LogIntoClusterDynamic()
	if err != nil {
		return nil, WorkbenchOutput{}, err
	}

	stopped, err := IsWorkbenchStopped(ctx, dyn, input.Namespace, input.WorkbenchName)
	if err != nil {
		return nil, WorkbenchOutput{}, err
	}
	if (input.Status == Stopped && stopped) || (input.Status == Running && !stopped) {
		return nil, WorkbenchOutput{Message: fmt.Sprintf("Workbench %s is already %s", input.WorkbenchName, input.Status)}, nil
	}

	patch := map[string]interface{}{}
	annotations := map[string]interface{}{}
	if input.Status == Stopped {
		annotations["kubeflow-resource-stopped"] = time.Now().UTC().Format(time.RFC3339)
	} else {
		annotations["kubeflow-resource-stopped"] = nil
	}
	patch["metadata"] = map[string]interface{}{
		"annotations": annotations,
	}

	patchBytes, err := json.Marshal(patch)
	if err != nil {
		return nil, WorkbenchOutput{}, fmt.Errorf("failed to marshal patch: %v", err)
	}

	_, err = dyn.Resource(workbenchesGVR).Namespace(input.Namespace).Patch(
		ctx,
		input.WorkbenchName,
		k8stypes.MergePatchType,
		patchBytes,
		metav1.PatchOptions{},
	)
	if err != nil {
		return nil, WorkbenchOutput{}, fmt.Errorf("failed to %s workbench %s: %v", input.Status, input.WorkbenchName, err)
	}

	return nil, WorkbenchOutput{Message: fmt.Sprintf("Workbench %s is %s", input.WorkbenchName, input.Status)}, nil
}

func CreateWorkbench(ctx context.Context, req *mcp.CallToolRequest, input CreateWorkbenchInput) (*mcp.CallToolResult, WorkbenchOutput, error) {
	return nil, WorkbenchOutput{Message: fmt.Sprintf("Workbench %s created", input.WorkbenchName)}, nil
}

func ListImages(ctx context.Context, req *mcp.CallToolRequest, input ListWorkbenchesInput) (*mcp.CallToolResult, ListImagesOutput, error) {
	dyn, err := LogIntoClusterDynamic()
	if err != nil {
		return nil, ListImagesOutput{}, err
	}

	imagesGVR := schema.GroupVersionResource{Group: "image.openshift.io", Version: "v1", Resource: "imagestreams"}
	images, err := dyn.Resource(imagesGVR).Namespace(input.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: "opendatahub.io/notebook-image=true",
	})
	if err != nil {
		return nil, ListImagesOutput{}, fmt.Errorf("failed to list images: %v", err)
	}

	msg := ""
	for _, image := range images.Items {
		msg += fmt.Sprintf("- %s\n", image.GetName())
	}
	return nil, ListImagesOutput{Images: msg}, nil
}
