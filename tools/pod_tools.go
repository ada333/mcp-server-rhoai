package tools

import (
	"context"
	"fmt"

	core "github.com/amaly/mcp-server-rhoai/core"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ListPods(ctx context.Context, req *mcp.CallToolRequest, input core.ListWorkbenchesInput) (*mcp.CallToolResult, core.ListPodsOutput, error) {
	clientset, err := GetClientSet()
	if err != nil {
		return nil, core.ListPodsOutput{}, err
	}

	pods, err := clientset.CoreV1().Pods(input.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, core.ListPodsOutput{}, fmt.Errorf("failed to list pods: %v", err)
	}

	var podInfos []core.PodInfo
	for _, pod := range pods.Items {
		podInfos = append(podInfos, core.PodInfo{
			Name:   pod.Name,
			Status: string(pod.Status.Phase),
		})
	}
	return nil, core.ListPodsOutput{Pods: podInfos}, nil
}
