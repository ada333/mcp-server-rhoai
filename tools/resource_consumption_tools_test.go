package tools

import (
	"context"
	"testing"

	core "github.com/amaly/mcp-server-rhoai/core"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func mockListWorkbenches(workbenches []core.WorkbenchInfo) func() {
	origList := listWorkbenchesFn

	listWorkbenchesFn = func(ctx context.Context, req *mcp.CallToolRequest, input core.ListWorkbenchesInput) (*mcp.CallToolResult, core.ListWorkbenchesResult, error) {
		var filtered []core.WorkbenchInfo
		for _, wb := range workbenches {
			if input.Namespace == "" || wb.Namespace == input.Namespace {
				filtered = append(filtered, wb)
			}
		}
		return nil, core.ListWorkbenchesResult{Workbenches: filtered}, nil
	}

	return func() {
		listWorkbenchesFn = origList
	}
}

func TestParseResourceValue(t *testing.T) {
	tests := []struct {
		input     string
		wantVal   float64
		wantUnit  string
		wantError bool
	}{
		{"2", 2, "", false},
		{"4Gi", 4, "Gi", false},
		{"500m", 500, "m", false},
		{"1.5Gi", 1.5, "Gi", false},
		{"0", 0, "", false},
		{"", 0, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			val, unit, err := parseResourceValue(tt.input)
			if (err != nil) != tt.wantError {
				t.Fatalf("parseResourceValue(%q) error = %v, wantError %v", tt.input, err, tt.wantError)
			}
			if val != tt.wantVal {
				t.Errorf("parseResourceValue(%q) val = %v, want %v", tt.input, val, tt.wantVal)
			}
			if unit != tt.wantUnit {
				t.Errorf("parseResourceValue(%q) unit = %q, want %q", tt.input, unit, tt.wantUnit)
			}
		})
	}
}

func TestSumResourceValues(t *testing.T) {
	tests := []struct {
		name   string
		values []string
		want   string
	}{
		{"empty", []string{}, "0"},
		{"single", []string{"4Gi"}, "4Gi"},
		{"multiple same unit", []string{"2Gi", "4Gi", "8Gi"}, "14Gi"},
		{"integers", []string{"1", "2", "3"}, "6"},
		{"with zeros", []string{"0", "2Gi", "0", "3Gi"}, "5Gi"},
		{"all zeros", []string{"0", "0", ""}, "0"},
		{"decimals", []string{"1.5", "2.5"}, "4"},
		{"millicores", []string{"500m", "300m"}, "800m"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sumResourceValues(tt.values)
			if got != tt.want {
				t.Errorf("sumResourceValues(%v) = %q, want %q", tt.values, got, tt.want)
			}
		})
	}
}

func TestAggregateResourceConsumption(t *testing.T) {
	workbenches := []core.WorkbenchInfo{
		{CPUUsage: "2", MemoryUsage: "4Gi", DiskUsage: "10Gi", GPUUsage: "1"},
		{CPUUsage: "4", MemoryUsage: "8Gi", DiskUsage: "20Gi", GPUUsage: "2"},
		{CPUUsage: "1", MemoryUsage: "2Gi", DiskUsage: "5Gi", GPUUsage: "0"},
	}

	result := aggregateResourceConsumption(workbenches)

	if result.CPUUsage != "7" {
		t.Errorf("expected CPU 7, got %q", result.CPUUsage)
	}
	if result.MemoryUsage != "14Gi" {
		t.Errorf("expected memory 14Gi, got %q", result.MemoryUsage)
	}
	if result.DiskUsage != "35Gi" {
		t.Errorf("expected disk 35Gi, got %q", result.DiskUsage)
	}
	if result.GPUUsage != "3" {
		t.Errorf("expected GPU 3, got %q", result.GPUUsage)
	}
}

func TestAggregateResourceConsumption_Empty(t *testing.T) {
	result := aggregateResourceConsumption(nil)

	if result.CPUUsage != "0" {
		t.Errorf("expected CPU 0, got %q", result.CPUUsage)
	}
	if result.MemoryUsage != "0" {
		t.Errorf("expected memory 0, got %q", result.MemoryUsage)
	}
}

func TestAggregateResourceConsumption_MixedEmptyValues(t *testing.T) {
	workbenches := []core.WorkbenchInfo{
		{CPUUsage: "2", MemoryUsage: "4Gi", DiskUsage: "", GPUUsage: ""},
		{CPUUsage: "", MemoryUsage: "2Gi", DiskUsage: "10Gi", GPUUsage: ""},
	}

	result := aggregateResourceConsumption(workbenches)

	if result.CPUUsage != "2" {
		t.Errorf("expected CPU 2, got %q", result.CPUUsage)
	}
	if result.MemoryUsage != "6Gi" {
		t.Errorf("expected memory 6Gi, got %q", result.MemoryUsage)
	}
	if result.DiskUsage != "10Gi" {
		t.Errorf("expected disk 10Gi, got %q", result.DiskUsage)
	}
	if result.GPUUsage != "0" {
		t.Errorf("expected GPU 0, got %q", result.GPUUsage)
	}
}

func TestListResourceConsumptionPerWorkbench(t *testing.T) {
	restore := mockListWorkbenches([]core.WorkbenchInfo{
		{Name: "wb-1", Namespace: "ns1", CPUUsage: "2", MemoryUsage: "4Gi", DiskUsage: "10Gi", GPUUsage: "1"},
		{Name: "wb-2", Namespace: "ns1", CPUUsage: "4", MemoryUsage: "8Gi", DiskUsage: "20Gi", GPUUsage: "0"},
	})
	defer restore()

	_, out, err := ListResourceConsumptionPerWorkbench(context.Background(), nil, core.ListResourceConsumptionPerWorkbenchInput{
		Namespace:     "ns1",
		WorkbenchName: "wb-1",
	})
	if err != nil {
		t.Fatalf("ListResourceConsumptionPerWorkbench returned error: %v", err)
	}
	if out.CPUUsage != "2" {
		t.Errorf("expected CPU 2, got %q", out.CPUUsage)
	}
	if out.MemoryUsage != "4Gi" {
		t.Errorf("expected memory 4Gi, got %q", out.MemoryUsage)
	}
	if out.GPUUsage != "1" {
		t.Errorf("expected GPU 1, got %q", out.GPUUsage)
	}
}

func TestListResourceConsumptionPerWorkbench_NotFound(t *testing.T) {
	restore := mockListWorkbenches([]core.WorkbenchInfo{
		{Name: "wb-1", Namespace: "ns1"},
	})
	defer restore()

	_, _, err := ListResourceConsumptionPerWorkbench(context.Background(), nil, core.ListResourceConsumptionPerWorkbenchInput{
		Namespace:     "ns1",
		WorkbenchName: "nonexistent",
	})
	if err == nil {
		t.Fatal("expected error for nonexistent workbench, got nil")
	}
}

func TestListResourceConsumptionPerNamespace(t *testing.T) {
	restore := mockListWorkbenches([]core.WorkbenchInfo{
		{Name: "wb-1", Namespace: "ns1", CPUUsage: "2", MemoryUsage: "4Gi", DiskUsage: "10Gi", GPUUsage: "1"},
		{Name: "wb-2", Namespace: "ns1", CPUUsage: "3", MemoryUsage: "6Gi", DiskUsage: "20Gi", GPUUsage: "0"},
		{Name: "wb-3", Namespace: "ns2", CPUUsage: "8", MemoryUsage: "16Gi", DiskUsage: "50Gi", GPUUsage: "2"},
	})
	defer restore()

	_, out, err := ListResourceConsumptionPerNamespace(context.Background(), nil, core.ListResourceConsumptionPerNamespaceInput{
		Namespace: "ns1",
	})
	if err != nil {
		t.Fatalf("ListResourceConsumptionPerNamespace returned error: %v", err)
	}
	if out.CPUUsage != "5" {
		t.Errorf("expected CPU 5 (2+3), got %q", out.CPUUsage)
	}
	if out.MemoryUsage != "10Gi" {
		t.Errorf("expected memory 10Gi (4+6), got %q", out.MemoryUsage)
	}
}

func TestListResourceConsumptionPerUser(t *testing.T) {
	restore := mockListWorkbenches([]core.WorkbenchInfo{
		{Name: "wb-1", Namespace: "ns1", User: "alice", CPUUsage: "2", MemoryUsage: "4Gi", DiskUsage: "10Gi", GPUUsage: "1"},
		{Name: "wb-2", Namespace: "ns2", User: "alice", CPUUsage: "3", MemoryUsage: "6Gi", DiskUsage: "15Gi", GPUUsage: "0"},
		{Name: "wb-3", Namespace: "ns1", User: "bob", CPUUsage: "8", MemoryUsage: "16Gi", DiskUsage: "50Gi", GPUUsage: "2"},
	})
	defer restore()

	_, out, err := ListResourceConsumptionPerUser(context.Background(), nil, core.ListResourceConsumptionPerUserInput{
		User: "alice",
	})
	if err != nil {
		t.Fatalf("ListResourceConsumptionPerUser returned error: %v", err)
	}
	if out.CPUUsage != "5" {
		t.Errorf("expected CPU 5 (2+3), got %q", out.CPUUsage)
	}
	if out.MemoryUsage != "10Gi" {
		t.Errorf("expected memory 10Gi (4+6), got %q", out.MemoryUsage)
	}
	if out.GPUUsage != "1" {
		t.Errorf("expected GPU 1 (1+0), got %q", out.GPUUsage)
	}
}

func TestListResourceConsumptionPerCluster(t *testing.T) {
	restore := mockListWorkbenches([]core.WorkbenchInfo{
		{Name: "wb-1", Namespace: "ns1", CPUUsage: "2", MemoryUsage: "4Gi", DiskUsage: "10Gi", GPUUsage: "1"},
		{Name: "wb-2", Namespace: "ns2", CPUUsage: "3", MemoryUsage: "6Gi", DiskUsage: "20Gi", GPUUsage: "2"},
	})
	defer restore()

	_, out, err := ListResourceConsumptionPerCluster(context.Background(), nil, struct{}{})
	if err != nil {
		t.Fatalf("ListResourceConsumptionPerCluster returned error: %v", err)
	}
	if out.CPUUsage != "5" {
		t.Errorf("expected CPU 5 (2+3), got %q", out.CPUUsage)
	}
	if out.MemoryUsage != "10Gi" {
		t.Errorf("expected memory 10Gi (4+6), got %q", out.MemoryUsage)
	}
	if out.DiskUsage != "30Gi" {
		t.Errorf("expected disk 30Gi (10+20), got %q", out.DiskUsage)
	}
	if out.GPUUsage != "3" {
		t.Errorf("expected GPU 3 (1+2), got %q", out.GPUUsage)
	}
}
