package main

type PodsOutput struct {
	Pods string `json:"pods" jsonschema_description:"the list of pods"`
}

type WorkbenchesOutput struct {
	Workbenches string `json:"workbenches" jsonschema_description:"the list of workbenches"`
}

type WorkBenchesInput struct {
	Namespace string `json:"namespace" jsonschema_description:"the namespace of the workbench"`
}

type WorkbenchToggleInput struct {
	Namespace     string `json:"namespace" jsonschema_description:"the namespace of the workbench"`
	WorkbenchName string `json:"workbenchName" jsonschema_description:"the name of the workbench (Notebook)"`
	Disable       bool   `json:"disable" jsonschema_description:"set true to disable (stop), false to enable (start)"`
}
