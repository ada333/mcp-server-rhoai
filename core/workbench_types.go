package core

type WorkbenchStatus int

const (
	Running WorkbenchStatus = iota
	Stopped
)

// used for printing the status
func (s WorkbenchStatus) String() string {
	switch s {
	case Running:
		return "running"
	case Stopped:
		return "stopped"
	default:
		return "unknown"
	}
}

type WorkbenchInfo struct {
	Name             string `json:"name" jsonschema:"the name of the workbench"`
	User             string `json:"user" jsonschema:"the user of the workbench"`
	Status           string `json:"status" jsonschema:"the status of the workbench"`
	ImageDisplayName string `json:"image" jsonschema:"the image of the workbench"`
	ImageTag         string `json:"imageTag" jsonschema:"the image tag of the workbench"`
	HardwareProfile  string `json:"hardwareProfile" jsonschema:"the name of the hardware profile of the workbench"`
	PVCName          string `json:"pvcName" jsonschema:"the name of the PVC of the workbench"`
	Namespace        string `json:"namespace" jsonschema:"the namespace of the workbench"`
	Uptime           string `json:"uptime" jsonschema:"the uptime of the workbench"`
	CPUUsage         string `json:"cpuUsage" jsonschema:"the CPU usage of the workbench"`
	MemoryUsage      string `json:"memoryUsage" jsonschema:"the memory usage of the workbench"`
	DiskUsage        string `json:"diskUsage" jsonschema:"the disk usage of the workbench"`
	GPUUsage         string `json:"gpuUsage" jsonschema:"the GPU usage of the workbench"`
}

type ListWorkbenchesResult struct {
	Workbenches []WorkbenchInfo `json:"workbenches" jsonschema:"the list of workbenches"`
}

type ListWorkbenchesInput struct {
	Namespace string `json:"namespace,omitempty" jsonschema:"the namespace to list workbenches from. If empty or not provided, lists workbenches across all namespaces."`
}

type ChangeWorkbenchStatusInput struct {
	Namespace     string          `json:"namespace" jsonschema:"the namespace of the workbench"`
	WorkbenchName string          `json:"workbenchName" jsonschema:"the name of the workbench"`
	Status        WorkbenchStatus `json:"status" jsonschema:"the target status: 0 for running, 1 for stopped"`
}

type CreateWorkbenchInput struct {
	Namespace        string          `json:"namespace" jsonschema:"the namespace of the workbench"`
	WorkbenchName    string          `json:"workbenchName" jsonschema:"the name of the workbench"`
	ImageDisplayName string          `json:"imageDisplayName" jsonschema:"the image display name - e.g. Jupyter Data Science"`
	ImageTag         string          `json:"imageTag" jsonschema:"the image tag/version"`
	HardwareProfile  HardwareProfile `json:"hardwareProfile,omitempty" jsonschema:"optional - the hardware profile to use. If not provided, default profile is used."`
	PVCName          string          `json:"pvcName,omitempty" jsonschema:"optional - the name of the PVC. If not provided, a PVC is auto-created with the workbench name."`
}

type UpdateWorkbenchInput struct {
	Namespace        string          `json:"namespace" jsonschema:"the namespace of the workbench"`
	WorkbenchName    string          `json:"workbenchName" jsonschema:"the name of the workbench"`
	ImageDisplayName string          `json:"imageDisplayName,omitempty" jsonschema:"optional - new image display name"`
	ImageTag         string          `json:"imageTag,omitempty" jsonschema:"optional - new image tag/version"`
	HardwareProfile  HardwareProfile `json:"hardwareProfile,omitempty" jsonschema:"optional - new hardware profile"`
	PVCName          string          `json:"pvcName,omitempty" jsonschema:"optional - new PVC name"`
}

type DeleteWorkbenchInput struct {
	Namespace     string `json:"namespace" jsonschema:"the namespace of the workbench"`
	WorkbenchName string `json:"workbenchName" jsonschema:"the name of the workbench"`
}
