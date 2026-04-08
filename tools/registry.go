package tools

import "github.com/modelcontextprotocol/go-sdk/mcp"

func RegisterWriteTools(server *mcp.Server) {
	registerPodTools(server)
	registerNamespaceTools(server)
	registerWorkbenchTools(server)
	registerImageTools(server)
	registerHardwareProfileTools(server)
	registerStorageTools(server)
}

func RegisterReadOnlyTools(server *mcp.Server) {
	registerWorkbenchListingTools(server)
	registerImageListingTools(server)
	registerHardwareProfileListingTools(server)
	registerStorageListingTools(server)
	registerResourceConsumptionListingTools(server)
}

func registerPodTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_pods",
		Description: "list the pods in a namespace",
	}, ListPods)
}

func registerNamespaceTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_namespaces",
		Description: "list all namespaces in the cluster",
	}, ListNamespaces)
}

func registerWorkbenchTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_workbench",
		Description: "create a new workbench with given name, image and image URL in a given project namespace",
	}, CreateWorkbench)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_workbench",
		Description: "update a workbench with given name, image and image URL in a given project namespace",
	}, UpdateWorkbench)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_workbench",
		Description: "delete a workbench with given name in a given project namespace",
	}, DeleteWorkbench)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "change_workbench_status",
		Description: "change the status of a workbench with given name in a given project namespace",
	}, ChangeWorkbenchStatus)
}

func registerWorkbenchListingTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_workbenches",
		Description: "list the workbenches in a given project namespace",
	}, ListWorkbenches)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_all_workbenches",
		Description: "list the workbenches across all namespaces",
	}, ListAllWorkbenches)
}

func registerImageTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_custom_image",
		Description: "create a new custom notebook image",
	}, CreateCustomImage)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_image",
		Description: "update an image with given name, description and location",
	}, UpdateImage)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_image",
		Description: "delete an image with given name",
	}, DeleteImage)
}

func registerImageListingTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_images",
		Description: "list all available notebook images",
	}, ListImages)
}

func registerHardwareProfileTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_hardware_profile",
		Description: "create a hardware profile with given name and resources",
	}, CreateHardwareProfile)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_hardware_profile",
		Description: "update a hardware profile with given name and resources",
	}, UpdateHardwareProfile)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_hardware_profile",
		Description: "delete a hardware profile with given name",
	}, DeleteHardwareProfile)
}

func registerHardwareProfileListingTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_hardware_profiles",
		Description: "list the hardware profiles in a given project namespace",
	}, ListHardwareProfiles)
}

func registerStorageTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_pvc",
		Description: "create a persistent volume claim with given name and size in a given project namespace",
	}, CreatePVC)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_pvc",
		Description: "update a persistent volume claim with given name and size in a given project namespace",
	}, UpdatePVC)
}

func registerStorageListingTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_pvcs",
		Description: "list the persistent volume claims in a given project namespace",
	}, ListPVCs)
}

func registerResourceConsumptionListingTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_resource_consumption_per_workbench",
		Description: "list the resource consumption of given workbench in a given namespace",
	}, ListResourceConsumptionPerWorkbench)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_resource_consumption_per_namespace",
		Description: "list the resource consumption of all workbenches in a given namespace",
	}, ListResourceConsumptionPerNamespace)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_resource_consumption_per_user",
		Description: "list the resource consumption of all workbenches of a given user",
	}, ListResourceConsumptionPerUser)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_resource_consumption_per_cluster",
		Description: "list the resource consumption of all workbenches in the cluster",
	}, ListResourceConsumptionPerCluster)
}
