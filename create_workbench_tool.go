package main

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
)

func CreateWorkbench(ctx context.Context, req *mcp.CallToolRequest, input CreateWorkbenchInput) (*mcp.CallToolResult, WorkbenchOutput, error) {
	dyn, err := getDynamicClient()
	if err != nil {
		return nil, WorkbenchOutput{}, err
	}

	err = createPersistentVolumeClaim(ctx, dyn, input.Namespace, input.WorkbenchName, "10Gi")
	if err != nil {
		return nil, WorkbenchOutput{}, fmt.Errorf("failed to create PVC: %v", err)
	}

	notebookArgs := fmt.Sprintf(`--ServerApp.port=8888
                  --ServerApp.token=''
                  --ServerApp.password=''
                  --ServerApp.base_url=/notebook/%s/%s
                  --ServerApp.quit_button=False`, input.Namespace, input.WorkbenchName)

	imageFull := input.ImageURL
	if input.ImageTag != "" {
		imageFull = fmt.Sprintf("%s:%s", input.ImageURL, input.ImageTag)
	}

	notebook := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "kubeflow.org/v1",
			"kind":       "Notebook",
			"metadata": map[string]interface{}{
				"name":      input.WorkbenchName,
				"namespace": input.Namespace,
				"labels": map[string]interface{}{
					"app":                        input.WorkbenchName,
					"opendatahub.io/dashboard":   "true",
					"opendatahub.io/odh-managed": "true",
				},
				"annotations": map[string]interface{}{
					"openshift.io/display-name":                     input.WorkbenchName,
					"openshift.io/description":                      "Created via MCP",
					"notebooks.opendatahub.io/inject-auth":          "true",
					"notebooks.opendatahub.io/last-image-selection": input.Image,
					"opendatahub.io/hardware-profile-name":          "default-profile",
					"opendatahub.io/hardware-profile-namespace":     "redhat-ods-applications",
				},
			},
			"spec": map[string]interface{}{
				"template": map[string]interface{}{
					"spec": map[string]interface{}{
						"serviceAccountName": "default",
						"enableServiceLinks": false,
						"containers": []interface{}{
							map[string]interface{}{
								"name":            input.WorkbenchName,
								"image":           imageFull,
								"imagePullPolicy": "Always",
								"workingDir":      "/opt/app-root/src",
								"ports": []interface{}{
									map[string]interface{}{
										"containerPort": 8888,
										"name":          "notebook-port",
										"protocol":      "TCP",
									},
								},
								"env": []interface{}{
									map[string]interface{}{
										"name":  "NOTEBOOK_ARGS",
										"value": notebookArgs,
									},
									map[string]interface{}{
										"name":  "JUPYTER_IMAGE",
										"value": input.ImageURL,
									},
								},
								"resources": map[string]interface{}{
									"limits": map[string]interface{}{
										"cpu":    "2",
										"memory": "4Gi",
									},
									"requests": map[string]interface{}{
										"cpu":    "2",
										"memory": "4Gi",
									},
								},
								"volumeMounts": []interface{}{
									map[string]interface{}{
										"mountPath": "/opt/app-root/src/",
										"name":      "storage-volume",
									},
									map[string]interface{}{
										"mountPath": "/dev/shm",
										"name":      "shm",
									},
								},
							},
						},
						"volumes": []interface{}{
							map[string]interface{}{
								"name": "storage-volume",
								"persistentVolumeClaim": map[string]interface{}{
									"claimName": input.WorkbenchName,
								},
							},
							map[string]interface{}{
								"name": "shm",
								"emptyDir": map[string]interface{}{
									"medium": "Memory",
								},
							},
						},
					},
				},
			},
		},
	}

	_, err = dyn.Resource(workbenchesGVR).Namespace(input.Namespace).Create(ctx, notebook, metav1.CreateOptions{})
	if err != nil {
		return nil, WorkbenchOutput{}, fmt.Errorf("failed to create notebook: %v", err)
	}

	return nil, WorkbenchOutput{Message: "Workbench was succesfully created!"}, nil
}

func createPersistentVolumeClaim(ctx context.Context, dyn dynamic.Interface, namespace, name, size string) error {
	pvc := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "PersistentVolumeClaim",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
				"labels": map[string]interface{}{
					"opendatahub.io/dashboard": "true",
				},
			},
			"spec": map[string]interface{}{
				"accessModes": []interface{}{"ReadWriteOnce"},
				"resources": map[string]interface{}{
					"requests": map[string]interface{}{
						"storage": size,
					},
				},
			},
		},
	}

	_, err := dyn.Resource(pvcGVR).Namespace(namespace).Create(ctx, pvc, metav1.CreateOptions{})
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}
	return nil
}
