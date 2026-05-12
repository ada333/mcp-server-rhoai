package resources

import "github.com/modelcontextprotocol/go-sdk/mcp"

func RegisterAllResources(server *mcp.Server) {
	server.AddResource(&mcp.Resource{
		URI:         "resource://mcp-server-rhoai/images",
		Name:        "Image Catalog",
		Description: "List of available notebook images their URLs and tags",
		MIMEType:    "application/json",
	}, ImagesResourceHandler)

	server.AddResource(&mcp.Resource{
		URI:         "resource://mcp-server-rhoai/hardware-profiles",
		Name:        "Hardware Profiles",
		Description: "List of available hardware profiles",
		MIMEType:    "application/json",
	}, DefaultHardwareResourceHandler)

	server.AddResource(&mcp.Resource{
		URI:         "resource://mcp-server-rhoai/namespaces",
		Name:        "Namespaces",
		Description: "List of all namespaces in the cluster",
		MIMEType:    "application/json",
	}, NamespacesResourceHandler)
}
