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
		Description: "create a new workbench. Requires namespace, workbenchName, imageDisplayName, and imageTag. PVC and hardware profile are optional - if not provided, a PVC is auto-created with the workbench name and default hardware profile is used.",
	}, CreateWorkbench)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_workbench",
		Description: "update a workbench. Requires namespace and workbenchName. Optionally update imageDisplayName, imageTag, hardwareProfile, or pvcName - only provide the fields you want to change.",
	}, UpdateWorkbench)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_workbench",
		Description: "delete a workbench with given name in a given project namespace",
	}, DeleteWorkbench)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "change_workbench_status",
		Description: "start or stop a workbench. Use status=0 to start (running), status=1 to stop (stopped).",
	}, ChangeWorkbenchStatus)
}

func registerWorkbenchListingTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_workbenches",
		Description: "list workbenches. If namespace is provided, lists workbenches in that namespace. If namespace is empty or not provided, lists all workbenches across all namespaces.",
	}, ListWorkbenches)
}

func registerImageTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_custom_image",
		Description: "create a custom notebook image. Requires imageName, imageLocation (docker URL), and imageDescription.",
	}, CreateCustomImage)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_image",
		Description: "update a custom image. Requires imageName. Optionally provide newImageName to rename or imageDescription to update description.",
	}, UpdateImage)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_image",
		Description: "delete a custom image by name",
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
		Description: "create a hardware profile. Requires hardwareProfileName and resources array. Each resource needs: resourceName, resourceIdentifier (e.g. 'cpu', 'memory'), resourceType, defaultCount, minCount, maxCount.",
	}, CreateHardwareProfile)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_hardware_profile",
		Description: "update a hardware profile. Requires hardwareProfileName and resources to update. Optionally provide newHardwareProfileName to rename.",
	}, UpdateHardwareProfile)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_hardware_profile",
		Description: "delete a hardware profile by name",
	}, DeleteHardwareProfile)
}

func registerHardwareProfileListingTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_hardware_profiles",
		Description: "list all available hardware profiles",
	}, ListHardwareProfiles)
}

func registerStorageTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_pvc",
		Description: "create a persistent volume claim. Requires namespace, pvcName, and size (e.g. '10Gi', '20Gi').",
	}, CreatePVC)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_pvc",
		Description: "update a persistent volume claim. Requires namespace and pvcName. Optionally provide size to resize (can only increase) or newPVCName to rename.",
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
