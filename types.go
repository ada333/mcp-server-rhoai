package main

import "k8s.io/apimachinery/pkg/runtime/schema"

var workbenchesGVR = schema.GroupVersionResource{Group: "kubeflow.org", Version: "v1", Resource: "notebooks"}

var pvcGVR = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "persistentvolumeclaims"}

type PodsOutput struct {
	Pods string `json:"pods" jsonschema_description:"the list of pods"`
}

type ListWorkbenchesResult struct {
	Workbenches string `json:"workbenches" jsonschema_description:"the list of workbenches"`
}

type ListWorkbenchesInput struct {
	Namespace string `json:"namespace" jsonschema_description:"the namespace of the workbench"`
}

type ChangeWorkbenchStatusInput struct {
	Namespace     string          `json:"namespace" jsonschema_description:"the namespace of the workbench"`
	WorkbenchName string          `json:"workbenchName" jsonschema_description:"the name of the workbench"`
	Status        WorkbenchStatus `json:"status" jsonschema_description:"the status of the workbench"`
}

type WorkbenchOutput struct {
	Message string `json:"message" jsonschema_description:"the message with result of workbench change"`
}

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

type CreateWorkbenchInput struct {
	Namespace        string `json:"namespace" jsonschema_description:"the namespace of the workbench"`
	WorkbenchName    string `json:"workbenchName" jsonschema_description:"the name of the workbench"`
	ImageDisplayName string `json:"imageDisplayName" jsonschema_description:"the image display name - f.e. Jupyter | Data Science | CPU | Python 3.12"`
	ImageTag         string `json:"imageTag" jsonschema_description:"the image tag "`
}

type ListImagesOutput struct {
	Images string `json:"images" jsonschema_description:"the list of images"`
}
