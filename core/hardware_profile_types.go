package core

type HardwareProfileOutput struct {
	Message         string `json:"message" jsonschema_description:"the message with result of hardware profile creation"`
	HardwareProfile string `json:"hardwareProfile" jsonschema_description:"the hardware profile created"`
}

type ListHardwareProfilesOutput struct {
	HardwareProfiles []HardwareProfile `json:"hardwareProfiles" jsonschema_description:"the list of hardware profiles"`
}

type HardwareProfile struct {
	HardwareProfileName string                    `json:"hardwareProfileName" jsonschema_description:"the name of the hardware profile"`
	Resources           []HardwareProfileResource `json:"resources" jsonschema_description:"the resources of the hardware profile"`
}

type HardwareProfileResource struct {
	ResourceName       string `json:"resourceName" jsonschema_description:"the name of the resource"`
	ResourceIdentifier string `json:"resourceIdentifier" jsonschema_description:"the identifier of the resource"`
	ResourceType       string `json:"resourceType" jsonschema_description:"the type of the resource"`
	DefaultCount       string `json:"defaultCount" jsonschema_description:"the default count of the resource"`
	MaxCount           string `json:"maxCount" jsonschema_description:"the max count of the resource"`
	MinCount           string `json:"minCount" jsonschema_description:"the min count of the resource"`
}

type UpdateHardwareProfileInput struct {
	HardwareProfileName    string                    `json:"hardwareProfileName" jsonschema_description:"the name of the hardware profile to update"`
	NewHardwareProfileName string                    `json:"newHardwareProfileName,omitempty" jsonschema_description:"optional - new name to rename the hardware profile to"`
	Resources              []HardwareProfileResource `json:"resources,omitempty" jsonschema_description:"optional - resources to update/add (merged with existing)"`
}

type DeleteHardwareProfileInput struct {
	HardwareProfileName string `json:"hardwareProfileName" jsonschema_description:"the name of the hardware profile"`
}
