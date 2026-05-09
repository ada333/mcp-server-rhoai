package core

type HardwareProfileOutput struct {
	Message         string `json:"message" jsonschema:"the message with result of hardware profile creation"`
	HardwareProfile string `json:"hardwareProfile" jsonschema:"the hardware profile created"`
}

type ListHardwareProfilesOutput struct {
	HardwareProfiles []HardwareProfile `json:"hardwareProfiles" jsonschema:"the list of hardware profiles"`
}

type HardwareProfile struct {
	HardwareProfileName string                    `json:"hardwareProfileName" jsonschema:"the name of the hardware profile"`
	Resources           []HardwareProfileResource `json:"resources" jsonschema:"the resources of the hardware profile"`
}

type HardwareProfileResource struct {
	ResourceName       string `json:"resourceName" jsonschema:"the name of the resource"`
	ResourceIdentifier string `json:"resourceIdentifier" jsonschema:"the identifier of the resource"`
	ResourceType       string `json:"resourceType" jsonschema:"the type of the resource"`
	DefaultCount       string `json:"defaultCount" jsonschema:"the default count of the resource"`
	MaxCount           string `json:"maxCount" jsonschema:"the max count of the resource"`
	MinCount           string `json:"minCount" jsonschema:"the min count of the resource"`
}

type UpdateHardwareProfileInput struct {
	HardwareProfileName    string                    `json:"hardwareProfileName" jsonschema:"the name of the hardware profile to update"`
	NewHardwareProfileName string                    `json:"newHardwareProfileName,omitempty" jsonschema:"optional - new name to rename the hardware profile to"`
	Resources              []HardwareProfileResource `json:"resources,omitempty" jsonschema:"optional - resources to update/add (merged with existing)"`
}

type DeleteHardwareProfileInput struct {
	HardwareProfileName string `json:"hardwareProfileName" jsonschema:"the name of the hardware profile"`
}
