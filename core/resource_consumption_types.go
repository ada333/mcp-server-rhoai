package core

type ListResourceConsumptionOutput struct {
	CPUUsage    string `json:"cpuUsage" jsonschema:"the CPU usage"`
	MemoryUsage string `json:"memoryUsage" jsonschema:"the memory usage"`
	DiskUsage   string `json:"diskUsage" jsonschema:"the disk usage"`
	GPUUsage    string `json:"gpuUsage" jsonschema:"the GPU usage"`
	UpTime      string `json:"upTime" jsonschema:"the up time"`
}

type ListResourceConsumptionPerWorkbenchInput struct {
	Namespace     string `json:"namespace" jsonschema:"the namespace of the workbench"`
	WorkbenchName string `json:"workbenchName" jsonschema:"the name of the workbench"`
}

type ListResourceConsumptionPerNamespaceInput struct {
	Namespace string `json:"namespace" jsonschema:"the namespace of the namespace"`
}

type ListResourceConsumptionPerUserInput struct {
	User string `json:"user" jsonschema:"the user of the user"`
}
