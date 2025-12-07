package main

import (
	"context"
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "workbencheslist",
		Version: "v1.0.0",
	}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "List Pods",
		Description: "list the pods in a namespace",
	}, ListPods)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "List Workbenches",
		Description: "list the workbenches in a given project namespace",
	}, ListWorkbenches)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "List All Workbenches",
		Description: "list the workbenches across all namespaces",
	}, ListAllWorkbenches)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "Change Workbench Status",
		Description: "change the status of a workbench with given namein a given project namespace",
	}, ChangeWorkbenchStatus)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "List Images",
		Description: "list the images in a given project namespace",
	}, ListImages)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "Create Workbench",
		Description: "create a new workbench with given name, image and image URL in a given project namespace",
	}, CreateWorkbench)

	server.AddResource(&mcp.Resource{
		URI:         "resource://mcp-test/images",
		Name:        "Image Catalog",
		Description: "List of available notebook images their URLs and tags",
		MIMEType:    "application/json",
	}, ImagesResourceHandler)

	server.AddPrompt(&mcp.Prompt{
		Name:        "create-workbench",
		Description: "Guide to create a workbench",
		Arguments: []*mcp.PromptArgument{
			{Name: "namespace", Description: "Target namespace", Required: true},
			{Name: "name", Description: "Workbench name", Required: true},
		},
	}, CreateWorkbenchPromptHandler)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}
