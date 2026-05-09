package core

type PodInfo struct {
	Name   string `json:"name" jsonschema:"the name of the pod"`
	Status string `json:"status" jsonschema:"the status/phase of the pod"`
}

type ListPodsOutput struct {
	Pods []PodInfo `json:"pods" jsonschema:"the list of pods"`
}

type DefaultToolOutput struct {
	Message string `json:"message" jsonschema:"the message with result of the tool execution"`
}

type ListNamespacesOutput struct {
	Namespaces []string `json:"namespaces" jsonschema:"the list of namespaces"`
}
