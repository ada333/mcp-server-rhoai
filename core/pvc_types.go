package core

type PVCInput struct {
	Namespace  string `json:"namespace" jsonschema_description:"the namespace of the PVC"`
	PVCName    string `json:"pvcName" jsonschema_description:"the name of the PVC"`
	NewPVCName string `json:"newPVCName,omitempty" jsonschema_description:"optional - new name to rename the PVC to"`
	Size       string `json:"size,omitempty" jsonschema_description:"the size of the PVC (e.g. '10Gi', '20Gi') - required for create, optional for update"`
}

type DeletePVCInput struct {
	Namespace string `json:"namespace" jsonschema_description:"the namespace of the PVC"`
	PVCName   string `json:"pvcName" jsonschema_description:"the name of the PVC"`
}

type PVCsOutput struct {
	PVCs string `json:"pvcs" jsonschema_description:"the list of PVCs"`
}

type ListPVCsInput struct {
	Namespace string `json:"namespace" jsonschema_description:"the namespace of the PVC"`
}
