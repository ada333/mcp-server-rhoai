package main

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

type ChangeWorkbenchStatusOutput struct {
	Message string `json:"message" jsonschema_description:"the message of the status change"`
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
