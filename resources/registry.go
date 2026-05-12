package resources

import "github.com/modelcontextprotocol/go-sdk/mcp"

func RegisterAllResources(server *mcp.Server) {
	server.AddResource(&mcp.Resource{
		URI:         "resource://mcp-server-rhoai/images",
		Name:        "Image Catalog",
		Description: "Returns all available notebook images with display names, URLs, versions/tags, and software dependencies. Use to find valid imageDisplayName and imageTag values for creating/updating workbenches.",
		MIMEType:    "application/json",
	}, ImagesResourceHandler)

	server.AddResource(&mcp.Resource{
		URI:         "resource://mcp-server-rhoai/hardware-profiles",
		Name:        "Hardware Profiles",
		Description: "Returns available hardware profiles with CPU, memory, and GPU resource limits. Use to find valid hardwareProfileName values for creating workbenches.",
		MIMEType:    "application/json",
	}, DefaultHardwareResourceHandler)

	server.AddResource(&mcp.Resource{
		URI:         "resource://mcp-server-rhoai/namespaces",
		Name:        "Namespaces",
		Description: "Returns all namespace names in the cluster. Use to find valid namespace values for workbench and PVC operations.",
		MIMEType:    "application/json",
	}, NamespacesResourceHandler)
}
