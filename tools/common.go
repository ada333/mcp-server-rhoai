package tools

import (
	"context"
	"fmt"

	core "github.com/amaly/mcp-server-rhoai/core"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

const defaultPVCSize = "10Gi"

var GetClientSet = func() (kubernetes.Interface, error) { return core.LogIntoClusterClientSet() }

var GetDynamicClient = func() (dynamic.Interface, error) { return core.LogIntoClusterDynamic() }

var getUptimeFromWorkbenchFn = getUptimeFromWorkbench
var getDiskUsageFromPVCFn = getDiskUsageFromPVC
var getImageInfoFn = func(ctx context.Context, displayName, version string) (string, string, string, error) {
	return GetImageInfo(ctx, displayName, version)
}

func convertToString(val interface{}) string {
	switch v := val.(type) {
	case string:
		return v
	case int:
		return fmt.Sprintf("%d", v)
	case int64:
		return fmt.Sprintf("%d", v)
	case float64:
		return fmt.Sprintf("%.0f", v)
	default:
		return ""
	}
}
