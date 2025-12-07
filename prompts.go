package main

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func CreateWorkbenchPromptHandler(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	var namespace, name string
	if req.Params.Arguments != nil {
		namespace = req.Params.Arguments["namespace"]
		name = req.Params.Arguments["name"]
	}

	return &mcp.GetPromptResult{
		Messages: []*mcp.PromptMessage{
			{
				Role: "user",
				Content: &mcp.TextContent{
					Text: fmt.Sprintf("I want to create a workbench named '%s' in namespace '%s'. Please check the 'Image Catalog' resource to find a suitable Data Science image (prefer Python 3.12) and then call the Create Workbench tool.", name, namespace),
				},
			},
		},
	}, nil
}
