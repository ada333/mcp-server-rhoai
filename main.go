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

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}
