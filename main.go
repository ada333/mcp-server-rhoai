package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

/*
type Input struct {
	Name string `json:"name" jsonschema_description:"the name of the person to greet"`
}

func SayHi(ctx context.Context, req *mcp.CallToolRequest, input Input) (*mcp.CallToolResult, string, error) {
	return nil, "Hi " + input.Name, nil
}

does it make sense to have this as a tool? or should user just oc login ...

	func ConnectToOpenShift(ctx context.Context, req *mcp.CallToolRequest, input Input) (*mcp.CallToolResult, string, error) {
		return nil, "Connected to OpenShift cluster", nil
	}
*/

type PodsOutput struct {
	Pods string `json:"pods" jsonschema_description:"the list of pods"`
}

type WorkbenchesOutput struct {
	Workbenches string `json:"workbenches" jsonschema_description:"the list of workbenches"`
}

func ListPods(ctx context.Context, req *mcp.CallToolRequest, input struct{ Namespace string }) (*mcp.CallToolResult, PodsOutput, error) {

	// logging in to cluster I think should be a separate tool for future
	kubeconfigPath := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, PodsOutput{}, fmt.Errorf("failed to load kubeconfig: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, PodsOutput{}, fmt.Errorf("failed to log into cluster: %v", err)
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

func ListWorkbenches(ctx context.Context, req *mcp.CallToolRequest, input struct{ Namespace string }) (*mcp.CallToolResult, WorkbenchesOutput, error) {

	// logging in to cluster I think should be a separate tool for future (dynamic or ClientSet options)
	kubeconfigPath := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, WorkbenchesOutput{}, fmt.Errorf("failed to load kubeconfig: %v", err)
	}
	dyn, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, WorkbenchesOutput{}, fmt.Errorf("failed to log into cluster: %v", err)
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

func main() {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "workbencheslist",
		Version: "v1.0.0",
	}, nil)
	/* testing tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "greet",
		Description: "say hi",
	}, SayHi)

		mcp.AddTool(server, &mcp.Tool{
			Name:        "List Pods",
			Description: "list the pods in a namespace",
		}, ListPods)
	*/
	mcp.AddTool(server, &mcp.Tool{
		Name:        "List Workbenches",
		Description: "list the workbenches in a given project namespace",
	}, ListWorkbenches)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}
