package tools

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/amaly/mcp-server-rhoai/core"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var listWorkbenchesFn = func(ctx context.Context, req *mcp.CallToolRequest, input core.ListWorkbenchesInput) (*mcp.CallToolResult, core.ListWorkbenchesResult, error) {
	return ListWorkbenches(ctx, req, input)
}

func parseResourceValue(s string) (float64, string, error) {
	if s == "" || s == "0" {
		return 0, "", nil
	}

	re := regexp.MustCompile(`^([\d.]+)(.*)$`)
	matches := re.FindStringSubmatch(s)
	if matches == nil {
		return 0, "", fmt.Errorf("invalid resource format: %s", s)
	}

	value, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0, "", fmt.Errorf("failed to parse number from %s: %w", s, err)
	}

	unit := strings.TrimSpace(matches[2])
	return value, unit, nil
}

// sumResourceValues sums multiple resource strings and returns the result with the (first) unit
func sumResourceValues(values []string) string {
	if len(values) == 0 {
		return "0"
	}

	var total float64
	var unit string

	for _, v := range values {
		if v == "" || v == "0" {
			continue
		}

		val, u, err := parseResourceValue(v)
		if err != nil {
			continue
		}

		total += val
		// if there would be multiple units, it wouldnt work correctly
		if unit == "" && u != "" {
			unit = u
		}
	}
	if total == 0 {
		return "0"
	}
	if total == float64(int64(total)) {
		return fmt.Sprintf("%d%s", int64(total), unit)
	}
	return fmt.Sprintf("%.2f%s", total, unit)
}

func ListResourceConsumptionPerWorkbench(ctx context.Context, req *mcp.CallToolRequest, input core.ListResourceConsumptionPerWorkbenchInput) (*mcp.CallToolResult, core.ListResourceConsumptionOutput, error) {
	_, workbenches, err := listWorkbenchesFn(ctx, req, core.ListWorkbenchesInput{Namespace: input.Namespace})
	if err != nil {
		return nil, core.ListResourceConsumptionOutput{}, err
	}
	// theoretically this is uneffective - there could be function that returns the workbench by name and namespace
	// and not list all workbenches (linear complexity)
	for _, wb := range workbenches.Workbenches {
		if wb.Name == input.WorkbenchName {
			return nil, core.ListResourceConsumptionOutput{
				CPUUsage:    wb.CPUUsage,
				MemoryUsage: wb.MemoryUsage,
				DiskUsage:   wb.DiskUsage,
				GPUUsage:    wb.GPUUsage,
			}, nil
		}
	}

	return nil, core.ListResourceConsumptionOutput{}, fmt.Errorf("workbench %s not found in namespace %s", input.WorkbenchName, input.Namespace)
}

func aggregateResourceConsumption(workbenches []core.WorkbenchInfo) core.ListResourceConsumptionOutput {
	var cpuValues, memoryValues, diskValues, gpuValues []string
	for _, wb := range workbenches {
		cpuValues = append(cpuValues, wb.CPUUsage)
		memoryValues = append(memoryValues, wb.MemoryUsage)
		diskValues = append(diskValues, wb.DiskUsage)
		gpuValues = append(gpuValues, wb.GPUUsage)
	}
	return core.ListResourceConsumptionOutput{
		CPUUsage:    sumResourceValues(cpuValues),
		MemoryUsage: sumResourceValues(memoryValues),
		DiskUsage:   sumResourceValues(diskValues),
		GPUUsage:    sumResourceValues(gpuValues),
	}
}

func ListResourceConsumptionPerNamespace(ctx context.Context, req *mcp.CallToolRequest, input core.ListResourceConsumptionPerNamespaceInput) (*mcp.CallToolResult, core.ListResourceConsumptionOutput, error) {
	_, workbenches, err := listWorkbenchesFn(ctx, req, core.ListWorkbenchesInput(input))
	if err != nil {
		return nil, core.ListResourceConsumptionOutput{}, err
	}
	return nil, aggregateResourceConsumption(workbenches.Workbenches), nil
}

func ListResourceConsumptionPerUser(ctx context.Context, req *mcp.CallToolRequest, input core.ListResourceConsumptionPerUserInput) (*mcp.CallToolResult, core.ListResourceConsumptionOutput, error) {
	_, workbenches, err := listWorkbenchesFn(ctx, req, core.ListWorkbenchesInput{})
	if err != nil {
		return nil, core.ListResourceConsumptionOutput{}, err
	}

	var userWorkbenches []core.WorkbenchInfo
	for _, wb := range workbenches.Workbenches {
		if wb.User == input.User {
			userWorkbenches = append(userWorkbenches, wb)
		}
	}
	return nil, aggregateResourceConsumption(userWorkbenches), nil
}

func ListResourceConsumptionPerCluster(ctx context.Context, req *mcp.CallToolRequest, input struct{}) (*mcp.CallToolResult, core.ListResourceConsumptionOutput, error) {
	_, workbenches, err := listWorkbenchesFn(ctx, req, core.ListWorkbenchesInput{})
	if err != nil {
		return nil, core.ListResourceConsumptionOutput{}, err
	}
	return nil, aggregateResourceConsumption(workbenches.Workbenches), nil
}
