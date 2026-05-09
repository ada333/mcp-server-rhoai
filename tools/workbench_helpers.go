package tools

import (
	"context"
	"fmt"
	"strings"

	core "github.com/amaly/mcp-server-rhoai/core"
	"github.com/amaly/mcp-server-rhoai/resources"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
)

const defaultPVCSize = "10Gi"

var getUptimeFromWorkbenchFn = getUptimeFromWorkbench
var getDiskUsageFromPVCFn = getDiskUsageFromPVC
var getImageInfoFn = GetImageInfo

func getPVCNameFromWorkbench(wb *unstructured.Unstructured) (string, error) {
	volumes, found, err := unstructured.NestedSlice(wb.Object, "spec", "template", "spec", "volumes")
	if err != nil {
		return "", fmt.Errorf("failed to get volumes: %v", err)
	}
	if !found {
		return "", nil
	}

	for _, vol := range volumes {
		volMap, ok := vol.(map[string]interface{})
		if !ok {
			continue
		}
		if pvc, found, _ := unstructured.NestedString(volMap, "persistentVolumeClaim", "claimName"); found {
			return pvc, nil
		}
	}
	return "", nil
}

func getResourceRequestsFromWorkbench(wb *unstructured.Unstructured) (cpuRequest, memoryRequest, gpuRequest string, err error) {
	containers, found, err := unstructured.NestedSlice(wb.Object, "spec", "template", "spec", "containers")
	if err != nil {
		return "", "", "", fmt.Errorf("failed to get containers: %v", err)
	}
	if !found || len(containers) == 0 {
		return "", "", "", nil
	}

	container, ok := containers[0].(map[string]interface{})
	if !ok {
		return "", "", "", nil
	}

	requests, found, err := unstructured.NestedStringMap(container, "resources", "requests")
	if err != nil {
		return "", "", "", fmt.Errorf("failed to get resource requests: %v", err)
	}
	if !found {
		return "", "", "", nil
	}

	cpuRequest = requests["cpu"]
	memoryRequest = requests["memory"]
	gpuRequest = requests["nvidia.com/gpu"] // FIX: maybe there can be other than nvidia.com/gpu?
	return cpuRequest, memoryRequest, gpuRequest, nil
}

func parseImageTag(annotations map[string]string) string {
	lastImageSelection := annotations["notebooks.opendatahub.io/last-image-selection"]
	if lastImageSelection == "" {
		return ""
	}
	parts := strings.Split(lastImageSelection, ":")
	if len(parts) > 1 {
		return parts[1]
	}
	return ""
}

func resolveWorkbenchStatus(annotations map[string]string) string {
	if annotations["kubeflow-resource-stopped"] != "" {
		return "stopped"
	}
	return "running"
}

func extractWorkbenchInfo(ctx context.Context, dyn dynamic.Interface, wb unstructured.Unstructured) (core.WorkbenchInfo, error) {
	name := wb.GetName()
	namespace := wb.GetNamespace()
	annotations := wb.GetAnnotations()

	status := resolveWorkbenchStatus(annotations)

	pvcName, err := getPVCNameFromWorkbench(&wb)
	if err != nil {
		return core.WorkbenchInfo{}, fmt.Errorf("failed to get PVC name for workbench %s: %v", name, err)
	}

	cpuRequest, memoryRequest, gpuRequest, err := getResourceRequestsFromWorkbench(&wb)
	if err != nil {
		return core.WorkbenchInfo{}, fmt.Errorf("failed to get resource requests for workbench %s: %v", name, err)
	}

	uptime := "0s"
	if status == "running" {
		uptime, err = getUptimeFromWorkbenchFn(name, namespace)
		if err != nil {
			return core.WorkbenchInfo{}, fmt.Errorf("failed to get uptime for workbench %s: %v", name, err)
		}
	}

	diskUsage := ""
	if pvcName != "" {
		diskUsage, err = getDiskUsageFromPVCFn(ctx, dyn, namespace, pvcName)
		if err != nil {
			return core.WorkbenchInfo{}, fmt.Errorf("failed to get disk usage for workbench %s: %v", name, err)
		}
	}

	return core.WorkbenchInfo{
		Name:             name,
		User:             annotations["opendatahub.io/username"],
		Status:           status,
		ImageDisplayName: annotations["opendatahub.io/image-display-name"],
		ImageTag:         parseImageTag(annotations),
		HardwareProfile:  annotations["opendatahub.io/hardware-profile-name"],
		PVCName:          pvcName,
		Namespace:        namespace,
		Uptime:           uptime,
		CPUUsage:         cpuRequest,
		MemoryUsage:      memoryRequest,
		DiskUsage:        diskUsage,
		GPUUsage:         gpuRequest,
	}, nil
}

func resolveHardwareProfile(input core.HardwareProfile) core.HardwareProfile {
	if input.HardwareProfileName != "" {
		return input
	}
	return resources.GetDefaultHardwareProfile()
}

func buildResourceRequirements(profile core.HardwareProfile) (limits, requests map[string]interface{}) {
	limits = make(map[string]interface{})
	requests = make(map[string]interface{})
	for _, r := range profile.Resources {
		limits[r.ResourceIdentifier] = r.MaxCount
		requests[r.ResourceIdentifier] = r.DefaultCount
	}
	return limits, requests
}

func resolveFullImageURL(repoURL, imageTag string) string {
	if imageTag != "" {
		return fmt.Sprintf("%s:%s", repoURL, imageTag)
	}
	return repoURL
}

func buildNotebookObject(input core.CreateWorkbenchInput, imageFull, imageName, gitCommit string, hardwareProfile core.HardwareProfile, limits, requests map[string]interface{}) *unstructured.Unstructured {
	notebookArgs := fmt.Sprintf(`--ServerApp.port=8888
                  --ServerApp.token=''
                  --ServerApp.password=''
                  --ServerApp.base_url=/notebook/%s/%s
                  --ServerApp.quit_button=False`, input.Namespace, input.WorkbenchName)

	return &unstructured.Unstructured{
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
					"opendatahub.io/image-display-name":                                input.ImageDisplayName,
					"openshift.io/display-name":                                        input.WorkbenchName,
					"openshift.io/description":                                         "Created via MCP",
					"notebooks.opendatahub.io/inject-auth":                             "true",
					"notebooks.opendatahub.io/last-image-selection":                    fmt.Sprintf("%s:%s", imageName, input.ImageTag),
					"notebooks.opendatahub.io/last-image-version-git-commit-selection": gitCommit,
					"opendatahub.io/hardware-profile-name":                             hardwareProfile.HardwareProfileName,
					"opendatahub.io/hardware-profile-namespace":                        core.GetDefaultNamespace(),
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
										"containerPort": int64(8888),
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
										"value": imageFull,
									},
								},
								"resources": map[string]interface{}{
									"limits":   limits,
									"requests": requests,
								},
								"volumeMounts": []interface{}{
									map[string]interface{}{
										"mountPath": "/opt/app-root/src/",
										"name":      input.PVCName,
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
								"name": input.PVCName,
								"persistentVolumeClaim": map[string]interface{}{
									"claimName": input.PVCName,
								},
							},
							map[string]interface{}{
								"name": "shm",
								"emptyDir": map[string]interface{}{
									"medium":    "Memory",
									"sizeLimit": "1Gi",
								},
							},
						},
					},
				},
			},
		},
	}
}

func updateWorkbenchImage(ctx context.Context, container map[string]interface{}, annotations map[string]string, displayName, imageTag string) error {
	repoURL, gitCommit, imageName, err := getImageInfoFn(ctx, displayName, imageTag)
	if err != nil {
		return fmt.Errorf("failed to lookup image info: %v", err)
	}
	imageFull := fmt.Sprintf("%s:%s", repoURL, imageTag)

	container["image"] = imageFull

	envs, _ := container["env"].([]interface{})
	for i, e := range envs {
		if envMap, ok := e.(map[string]interface{}); ok && envMap["name"] == "JUPYTER_IMAGE" {
			envMap["value"] = imageFull
			envs[i] = envMap
		}
	}
	container["env"] = envs

	annotations["opendatahub.io/image-display-name"] = displayName
	annotations["notebooks.opendatahub.io/last-image-selection"] = fmt.Sprintf("%s:%s", imageName, imageTag)
	annotations["notebooks.opendatahub.io/last-image-version-git-commit-selection"] = gitCommit
	return nil
}

func updateWorkbenchHardwareProfile(container map[string]interface{}, annotations map[string]string, profile core.HardwareProfile) {
	limits, requests := buildResourceRequirements(profile)
	container["resources"] = map[string]interface{}{
		"limits":   limits,
		"requests": requests,
	}
	annotations["opendatahub.io/hardware-profile-name"] = profile.HardwareProfileName
	annotations["opendatahub.io/hardware-profile-namespace"] = core.GetDefaultNamespace()
}

func updateWorkbenchPVC(container map[string]interface{}, workbench *unstructured.Unstructured, pvcName string) error {
	volumeMounts, _ := container["volumeMounts"].([]interface{})
	for i, vm := range volumeMounts {
		if vmMap, ok := vm.(map[string]interface{}); ok && vmMap["mountPath"] == "/opt/app-root/src/" {
			vmMap["name"] = pvcName
			volumeMounts[i] = vmMap
		}
	}
	container["volumeMounts"] = volumeMounts

	volumes, _, _ := unstructured.NestedSlice(workbench.Object, "spec", "template", "spec", "volumes")
	for i, v := range volumes {
		if vMap, ok := v.(map[string]interface{}); ok {
			if _, hasPVC := vMap["persistentVolumeClaim"]; hasPVC {
				vMap["name"] = pvcName
				if pvcMap, ok := vMap["persistentVolumeClaim"].(map[string]interface{}); ok {
					pvcMap["claimName"] = pvcName
					vMap["persistentVolumeClaim"] = pvcMap
				}
				volumes[i] = vMap
			}
		}
	}
	return unstructured.SetNestedSlice(workbench.Object, volumes, "spec", "template", "spec", "volumes")
}

func getFirstContainer(workbench *unstructured.Unstructured) (map[string]interface{}, []interface{}, error) {
	containers, _, _ := unstructured.NestedSlice(workbench.Object, "spec", "template", "spec", "containers")
	if len(containers) == 0 {
		return nil, nil, fmt.Errorf("workbench has no containers")
	}
	container, ok := containers[0].(map[string]interface{})
	if !ok {
		return nil, nil, fmt.Errorf("unexpected container format")
	}
	return container, containers, nil
}

func IsWorkbenchStopped(ctx context.Context, dyn dynamic.Interface, namespace, workbenchName string) (bool, error) {
	current, err := dyn.Resource(core.WorkbenchesGVR).Namespace(namespace).Get(ctx, workbenchName, metav1.GetOptions{})
	if err != nil {
		return false, fmt.Errorf("failed to get workbench %s: %v", workbenchName, err)
	}
	currentAnnotations := current.GetAnnotations()
	currentStopped := false
	if currentAnnotations != nil {
		if _, ok := currentAnnotations["kubeflow-resource-stopped"]; ok {
			currentStopped = true
		}
	}
	return currentStopped, nil
}

// GetImageInfo retrieves image information from display name and version
// Returns: repoURL, gitCommit, imageName, error
func GetImageInfo(ctx context.Context, displayName, version string) (string, string, string, error) {
	dyn, err := GetDynamicClient()
	if err != nil {
		return "", "", "", err
	}

	images, err := dyn.Resource(core.ImagesGVR).Namespace(core.GetDefaultNamespace()).List(ctx, metav1.ListOptions{
		LabelSelector: "opendatahub.io/notebook-image=true",
	})
	if err != nil {
		return "", "", "", fmt.Errorf("failed to list images: %v", err)
	}

	for _, image := range images.Items {
		annotations := image.GetAnnotations()
		if annotations["opendatahub.io/notebook-image-name"] == displayName {
			repoURL, found, err := unstructured.NestedString(image.Object, "status", "dockerImageRepository")
			if !found || err != nil {
				repoURL = "URL not available"
			}
			imageName := image.GetName()

			tagsRaw, _, _ := unstructured.NestedSlice(image.Object, "spec", "tags")
			for _, t := range tagsRaw {
				tagMap, ok := t.(map[string]interface{})
				if !ok {
					continue
				}
				tagName, _ := tagMap["name"].(string)
				if tagName == version {
					tagAnnotations, _, _ := unstructured.NestedStringMap(tagMap, "annotations")
					return repoURL, tagAnnotations["opendatahub.io/notebook-build-commit"], imageName, nil
				}
			}
		}
	}
	return "", "", "", fmt.Errorf("image not found: %s:%s", displayName, version)
}

func getUptimeFromWorkbench(workbenchName string, nameSpace string) (string, error) {
	ctx := context.Background()
	clientset, err := GetClientSet()
	if err != nil {
		return "", fmt.Errorf("failed to get client set: %v", err)
	}

	labelSelector := fmt.Sprintf("notebook-name=%s", workbenchName)
	pods, err := clientset.CoreV1().Pods(nameSpace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return "", fmt.Errorf("failed to list pods for workbench %s in namespace %s: %v", workbenchName, nameSpace, err)
	}

	if len(pods.Items) == 0 {
		return "", nil
	}

	pod := pods.Items[0]
	if pod.Status.StartTime == nil {
		return "", nil
	}
	return pod.Status.StartTime.String(), nil
}
