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
