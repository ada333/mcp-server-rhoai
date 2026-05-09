package core

type PVCInput struct {
	Namespace  string `json:"namespace" jsonschema:"the namespace of the PVC"`
	PVCName    string `json:"pvcName" jsonschema:"the name of the PVC"`
	NewPVCName string `json:"newPVCName,omitempty" jsonschema:"optional - new name to rename the PVC to"`
	Size       string `json:"size,omitempty" jsonschema:"the size of the PVC (e.g. '10Gi', '20Gi') - required for create, optional for update"`
}

type DeletePVCInput struct {
	Namespace string `json:"namespace" jsonschema:"the namespace of the PVC"`
	PVCName   string `json:"pvcName" jsonschema:"the name of the PVC"`
}

type PVCInfo struct {
	Name   string `json:"name" jsonschema:"the name of the PVC"`
	Size   string `json:"size" jsonschema:"the size of the PVC"`
	Status string `json:"status" jsonschema:"the status of the PVC"`
}

type ListPVCsOutput struct {
	PVCs []PVCInfo `json:"pvcs" jsonschema:"the list of PVCs"`
}

type ListPVCsInput struct {
	Namespace string `json:"namespace" jsonschema:"the namespace of the PVC"`
}
