package tools

import (
	"context"
	"fmt"
	"time"

	core "github.com/amaly/mcp-server-rhoai/core"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func ListHardwareProfiles(ctx context.Context, req *mcp.CallToolRequest, input struct{}) (*mcp.CallToolResult, core.ListHardwareProfilesOutput, error) {
	dyn, err := GetDynamicClient()
	if err != nil {
		return nil, core.ListHardwareProfilesOutput{}, err
	}

	hardwareProfiles, err := dyn.Resource(core.HardwareProfilesGVR).Namespace(core.GetDefaultNamespace()).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, core.ListHardwareProfilesOutput{}, fmt.Errorf("failed to list hardware profiles: %v", err)
	}

	var result []core.HardwareProfile
	for _, profile := range hardwareProfiles.Items {
		hwResources, err := GetResourcesFromHardwareProfile(&profile)
		if err != nil {
			continue
		}

		result = append(result, core.HardwareProfile{
			HardwareProfileName: profile.GetName(),
			Resources:           hwResources,
		})
	}
	return nil, core.ListHardwareProfilesOutput{HardwareProfiles: result}, nil
}

func CreateHardwareProfile(ctx context.Context, req *mcp.CallToolRequest, input core.HardwareProfile) (*mcp.CallToolResult, core.DefaultToolOutput, error) {
	dyn, err := GetDynamicClient()
	if err != nil {
		return nil, core.DefaultToolOutput{}, err
	}

	namespace := core.GetDefaultNamespace()

	identifiers := make([]interface{}, len(input.Resources))
	for i, res := range input.Resources {
		identifiers[i] = map[string]interface{}{
			"displayName":  res.ResourceName,
			"identifier":   res.ResourceIdentifier,
			"resourceType": res.ResourceType,
			"defaultCount": res.DefaultCount,
			"maxCount":     res.MaxCount,
			"minCount":     res.MinCount,
		}
	}

	annotations := map[string]interface{}{
		"opendatahub.io/dashboard-feature-visibility": "[]",
		"opendatahub.io/disabled":                     "false",
		"opendatahub.io/display-name":                 input.HardwareProfileName,
		"opendatahub.io/modified-date":                time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
	}

	hardwareProfile := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "infrastructure.opendatahub.io/v1",
			"kind":       "HardwareProfile",
			"metadata": map[string]interface{}{
				"name":        input.HardwareProfileName,
				"namespace":   namespace,
				"annotations": annotations,
			},
			"spec": map[string]interface{}{
				"identifiers": identifiers,
			},
		},
	}

	_, err = dyn.Resource(core.HardwareProfilesGVR).Namespace(namespace).Create(ctx, hardwareProfile, metav1.CreateOptions{})
	if err != nil {
		return nil, core.DefaultToolOutput{}, fmt.Errorf("failed to create hardware profile: %v", err)
	}

	return nil, core.DefaultToolOutput{
		Message: "Hardware Profile was successfully created!",
	}, nil
}

func UpdateHardwareProfile(ctx context.Context, req *mcp.CallToolRequest, input core.UpdateHardwareProfileInput) (*mcp.CallToolResult, core.DefaultToolOutput, error) {
	dyn, err := GetDynamicClient()
	if err != nil {
		return nil, core.DefaultToolOutput{}, err
	}

	hardwareProfile, err := dyn.Resource(core.HardwareProfilesGVR).Namespace(core.GetDefaultNamespace()).Get(ctx, input.HardwareProfileName, metav1.GetOptions{})
	if err != nil {
		return nil, core.DefaultToolOutput{}, fmt.Errorf("failed to get hardware profile: %v", err)
	}

	if input.NewHardwareProfileName != "" {
		hardwareProfile.SetAnnotations(map[string]string{
			"opendatahub.io/display-name": input.NewHardwareProfileName,
		})
	}

	if len(input.Resources) > 0 {
		existing, err := GetResourcesFromHardwareProfile(hardwareProfile)
		if err != nil {
			return nil, core.DefaultToolOutput{}, fmt.Errorf("failed to get resources from hardware profile: %v", err)
		}

		merged := make(map[string]core.HardwareProfileResource, len(existing))
		// resource identifier should be unique - so we overwrite existing one with the updated one
		for _, res := range existing {
			merged[res.ResourceIdentifier] = res
		}
		for _, res := range input.Resources {
			merged[res.ResourceIdentifier] = res
		}
		updatedIdentifiers := make([]interface{}, 0, len(merged))
		for _, res := range merged {
			updatedIdentifiers = append(updatedIdentifiers, map[string]interface{}{
				"displayName":  res.ResourceName,
				"identifier":   res.ResourceIdentifier,
				"resourceType": res.ResourceType,
				"defaultCount": res.DefaultCount,
				"maxCount":     res.MaxCount,
				"minCount":     res.MinCount,
			})
		}

		if err := unstructured.SetNestedSlice(hardwareProfile.Object, updatedIdentifiers, "spec", "identifiers"); err != nil {
			return nil, core.DefaultToolOutput{}, fmt.Errorf("failed to set hardware profile identifiers: %v", err)
		}
	}

	_, err = dyn.Resource(core.HardwareProfilesGVR).Namespace(core.GetDefaultNamespace()).Update(ctx, hardwareProfile, metav1.UpdateOptions{})
	if err != nil {
		return nil, core.DefaultToolOutput{}, fmt.Errorf("failed to update hardware profile: %v", err)
	}

	return nil, core.DefaultToolOutput{Message: "Hardware Profile was successfully updated!"}, nil
}

func DeleteHardwareProfile(ctx context.Context, req *mcp.CallToolRequest, input core.DeleteHardwareProfileInput) (*mcp.CallToolResult, core.DefaultToolOutput, error) {
	dyn, err := GetDynamicClient()
	if err != nil {
		return nil, core.DefaultToolOutput{}, err
	}

	err = dyn.Resource(core.HardwareProfilesGVR).Namespace(core.GetDefaultNamespace()).Delete(ctx, input.HardwareProfileName, metav1.DeleteOptions{})
	if err != nil {
		return nil, core.DefaultToolOutput{}, fmt.Errorf("failed to delete hardware profile %s: %v", input.HardwareProfileName, err)
	}

	return nil, core.DefaultToolOutput{Message: fmt.Sprintf("Hardware Profile %s was successfully deleted", input.HardwareProfileName)}, nil
}

func GetResourcesFromHardwareProfile(hardWareProfile *unstructured.Unstructured) ([]core.HardwareProfileResource, error) {
	identifiers, foundIdentifiers, err := unstructured.NestedSlice(hardWareProfile.Object, "spec", "identifiers")
	if !foundIdentifiers || err != nil {
		return nil, fmt.Errorf("failed to get identifiers: %v", err)
	}

	resources := make([]core.HardwareProfileResource, 0, len(identifiers))
	for _, identifier := range identifiers {
		identifierMap, okIdentifierMap := identifier.(map[string]interface{})
		if !okIdentifierMap {
			continue
		}

		resources = append(resources, core.HardwareProfileResource{
			ResourceName:       convertToString(identifierMap["displayName"]),
			ResourceIdentifier: convertToString(identifierMap["identifier"]),
			ResourceType:       convertToString(identifierMap["resourceType"]),
			DefaultCount:       convertToString(identifierMap["defaultCount"]),
			MaxCount:           convertToString(identifierMap["maxCount"]),
			MinCount:           convertToString(identifierMap["minCount"]),
		})
	}
	return resources, nil
}
