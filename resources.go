package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type ImageDef struct {
	Name     string   `json:"name"`
	URL      string   `json:"url"`
	Versions []string `json:"versions"`
}

func GetImages(ctx context.Context) ([]ImageDef, error) {
	dyn, err := getDynamicClient()
	if err != nil {
		return nil, err
	}

	imagesGVR := schema.GroupVersionResource{Group: "image.openshift.io", Version: "v1", Resource: "imagestreams"}
	images, err := dyn.Resource(imagesGVR).Namespace("redhat-ods-applications").List(ctx, metav1.ListOptions{
		LabelSelector: "opendatahub.io/notebook-image=true",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list images: %v", err)
	}

	var result []ImageDef
	for _, image := range images.Items {
		annotations := image.GetAnnotations()
		displayName := annotations["opendatahub.io/notebook-image-name"]

		repoURL, found, err := unstructured.NestedString(image.Object, "status", "dockerImageRepository")
		if !found || err != nil {
			repoURL = "URL not available"
		}

		tagsRaw, _, _ := unstructured.NestedSlice(image.Object, "spec", "tags")

		var versions []string
		for _, t := range tagsRaw {
			tagMap, ok := t.(map[string]interface{})
			if ok {
				tagName, _ := tagMap["name"].(string)
				versions = append(versions, tagName)
			}
		}

		result = append(result, ImageDef{
			Name:     displayName,
			URL:      repoURL,
			Versions: versions,
		})
	}
	return result, nil
}

func ImagesResourceHandler(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	images, err := GetImages(ctx)
	if err != nil {
		return nil, err
	}

	jsonBytes, err := json.Marshal(images)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal images: %v", err)
	}

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:      req.Params.URI,
				MIMEType: "application/json",
				Text:     string(jsonBytes),
			},
		},
	}, nil
}
