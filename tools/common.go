package tools

import (
	"fmt"

	core "github.com/amaly/mcp-server-rhoai/core"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

var GetClientSet = func() (kubernetes.Interface, error) { return core.LogIntoClusterClientSet() }

var GetDynamicClient = func() (dynamic.Interface, error) { return core.LogIntoClusterDynamic() }

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
