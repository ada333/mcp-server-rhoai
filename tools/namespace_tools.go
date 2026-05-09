package tools

import (
	"context"
	"fmt"

	core "github.com/amaly/mcp-server-rhoai/core"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ListNamespaces(ctx context.Context, req *mcp.CallToolRequest, input struct{}) (*mcp.CallToolResult, core.ListNamespacesOutput, error) {
	namespaces, err := GetAllNamespaces(ctx)
	if err != nil {
		return nil, core.ListNamespacesOutput{}, err
	}

	return nil, core.ListNamespacesOutput{Namespaces: namespaces}, nil
}

func GetAllNamespaces(ctx context.Context) ([]string, error) {
	clientset, err := GetClientSet()
	if err != nil {
		return nil, err
	}

	namespaces, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %v", err)
	}

	var names []string
	for _, ns := range namespaces.Items {
		names = append(names, ns.Name)
	}
	return names, nil
}
