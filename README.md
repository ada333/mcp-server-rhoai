# MCP Server for Red Hat OpenShift AI

This project implements a Model Context Protocol (MCP) server for managing Red Hat OpenShift AI (RHOAI) workbenches. It provides tools for listing, creating, and managing workbenches, images, hardware profiles, and storage in OpenShift clusters, also it provides tools for monitoring resource consumption by workbenches per user, workbench, namespace or cluster.
This project is part of a Bachelor's thesis.

## Prerequisites

- Go 1.24.0 or later
- Access to an OpenShift cluster with RHOAI installed
- `oc` CLI logged in to your cluster
- (Optional) golangci-lint for code linting

## Installation

1. Clone the repository:
```bash
git clone https://github.com/amaly/mcp-server-rhoai.git
cd mcp-server-rhoai
```

2. Download dependencies:
```bash
go mod download
```

3. Ensure you're logged in to your OpenShift cluster:
```bash
oc login <your-cluster-url>
```

## Building

The project includes a Makefile for easy building:

### Build with Make

```bash
# Build with linting and testing (recommended)
make build

# Build without running lint and tests (faster)
go build -o mcp-server-rhoai
```

The build process will:
1. Run linting checks (golangci-lint)
2. Execute all tests
3. Compile the binary as `mcp-server-rhoai`

### Build Outputs

- Binary location: `./mcp-server-rhoai`

## Testing

The project has a comprehensive unit test suite (60 tests across 7 test files) covering every tool module. Tests use Kubernetes fake clients (`dynamicfake.NewSimpleDynamicClient`) and function-variable mocking to isolate logic from live cluster dependencies, so no running cluster is needed.

Run tests using Make or directly with Go:

```bash
# Run all tests with Make
make test

# Run tests directly with verbose output
go test -v ./tools/... ./resources/... ./prompts/...
```

Test coverage includes:
- **Workbench tools** — CRUD operations, status transitions, and helper functions (image tag parsing, resource extraction, PVC name resolution)
- **Storage tools** — PVC creation, deletion, and size updates including rejection of size decreases
- **Image tools** — Custom image CRUD, default-image and in-use protection checks, image listing
- **Hardware profile tools** — Profile CRUD and resource identifier parsing
- **Resource consumption tools** — Value parsing (millicores, Gi suffixes, decimals), aggregation, and per-workbench/namespace/user/cluster queries
- **Namespace & pod tools** — Listing and filtering by namespace

Each tool module has a dedicated `*_test.go` file with both success and error-path cases. Table-driven subtests are used where multiple input variations need to be verified.

## Evaluation

Uses [Promptfoo](https://www.promptfoo.dev/) for testing tool selection accuracy.

### Run Evaluation

```bash
export GOOGLE_API_KEY=your-key-here
make eval        # run tests
make eval-view   # run and open results in browser
```

### What It Tests

- **Tool selection** - correct tool called for each prompt
- **Parameter extraction** - arguments passed correctly (namespace, workbench name, etc.)

### Key Learnings

Tool descriptions must be explicit for model-agnostic compatibility:

- Mark optional fields with `omitempty` and describe defaults in jsonschema descriptions
- Include valid values in descriptions (e.g., "status: 0 for running, 1 for stopped")  
- Simpler models (Gemini Flash) are stricter than Claude - good for testing robustness

## Configuration

The MCP server can be configured in AI assistants to enable workbench management through natural language.

### VSCode/Cursor Setup

1. **Build the server:**
```bash
make build
```

2. **Open MCP Settings in Cursor:**
   - Press `Ctrl + Shift + P` (Windows/Linux) or `Cmd + Shift + P` (Mac)
   - Type "Open MCP Settings"
   - Click "New MCP server"

3. **Add configuration:**

**Read-only mode** (default, safer):
```json
{
  "mcpServers": {
    "mcp-server-rhoai": {
      "command": "/home/amaly/mcp-server-rhoai/mcp-server-rhoai"
    }
  }
}
```

**Read-write mode** (full access):
```json
{
  "mcpServers": {
    "mcp-server-rhoai": {
      "command": "/home/amaly/mcp-server-rhoai/mcp-server-rhoai",
      "env": {
        "MCP_RHOAI_MODE": "write"
      }
    }
  }
}
```

4. **Verify installation:**
   - Check that the server is enabled in Cursor MCP settings
   - You should see available tools listed

5. **Start using:**
   - Ask the AI assistant: "List all workbenches in namespace mcp-test"
   - Or: "Show me resource consumption per namespace"

### Claude Desktop Setup

1. **Build the server:**
```bash
make build
```

2. **Locate Claude Desktop config file:**

   - **macOS:** `~/Library/Application Support/Claude/claude_desktop_config.json`
   - **Windows:** `%APPDATA%\Claude\claude_desktop_config.json`
   - **Linux:** `~/.config/Claude/claude_desktop_config.json`

3. **Edit the config file** and add the MCP server:

**Read-only mode** (default):
```json
{
  "mcpServers": {
    "mcp-server-rhoai": {
      "command": "/home/amaly/mcp-server-rhoai/mcp-server-rhoai"
    }
  }
}
```

**Read-write mode** (full access):
```json
{
  "mcpServers": {
    "mcp-server-rhoai": {
      "command": "/home/amaly/mcp-server-rhoai/mcp-server-rhoai",
      "env": {
        "MCP_RHOAI_MODE": "write"
      }
    }
  }
}
```

4. **Restart Claude Desktop** for changes to take effect

5. **Verify connection:**
   - Look for the MCP connection indicator in Claude Desktop
   - Ask Claude: "What workbenches are available in my cluster?"

### Important Notes

**Permissions:** The MCP server uses your current `oc` login credentials. All operations are performed with your user's permissions.

**Security:** Start with read-only mode to explore safely. Only enable write mode when you need to create or modify resources.

## Tool Modes

The server supports two operation modes for safety and flexibility:

### Read-Only Mode (Default)

Only listing and querying tools are available. This mode is **safer** and prevents accidental modifications to your cluster.

**Available operations:**
- List workbenches, images, hardware profiles, PVCs
- Query resource consumption
- View namespace and pod information

### Read-Write Mode

All tools are available, including create, delete, and modify operations.

**Additional operations:**
- Create/Delete workbenches
- Create/Delete custom images
- Create/Delete hardware profiles
- Create PVCs
- Start/Stop workbenches

### Configuring the Mode

The mode can be set via command-line flag or environment variable:

**Command-line flag (recommended for promptfoo/scripts):**
```bash
./mcp-server-rhoai -mode write
```

**Environment variable (for MCP client configs):**
```bash
export MCP_RHOAI_MODE=write
```

| Value | Mode | Description |
|-------|------|-------------|
| Not set | Read-only | Default, safe mode |
| `write` | Read-write | Full access to all tools |

The `-mode` flag takes precedence over the environment variable.

## Available Tools

### Read-Only Tools (Always Available)

- **List Pods** - List pods in a namespace
- **List Namespaces** - List all namespaces in the cluster
- **List Workbenches** - List workbenches in a specific namespace
- **List All Workbenches** - List workbenches across all namespaces
- **List Images** - List all available notebook images
- **List Hardware Profiles** - List available hardware profiles
- **List PVCs** - List persistent volume claims in a namespace
- **List Resource Consumption Per Workbench** - Get resource usage for a specific workbench
- **List Resource Consumption Per Namespace** - Get resource usage for all workbenches in a namespace
- **List Resource Consumption Per User** - Get resource usage for a specific user
- **List Resource Consumption Per Cluster** - Get cluster-wide resource usage

### Write Tools (Only in Write Mode)

- **Create Workbench** - Create a new workbench with specified configuration
- **Delete Workbench** - Delete an existing workbench
- **Change Workbench Status** - Start or stop a workbench
- **Create Custom Image** - Create a custom notebook image
- **Delete Image** - Delete a custom image
- **Create Hardware Profile** - Create a new hardware profile
- **Delete Hardware Profile** - Delete a hardware profile
- **Create PVC** - Create a persistent volume claim

## Development

### Project Structure

```
mcp-server-rhoai/
├── main.go                    # Entry point
├── core/                      # Shared types and utilities
│   ├── gvr.go                 # Kubernetes GroupVersionResource definitions
│   ├── common_types.go        # Shared output/namespace types
│   ├── workbench_types.go     # Workbench-related types
│   ├── image_types.go         # Image-related types
│   ├── hardware_profile_types.go  # Hardware profile types
│   ├── pvc_types.go           # PVC-related types
│   ├── resource_consumption_types.go  # Resource consumption types
│   └── logging.go             # Logging utilities
├── tools/                     # MCP tool implementations
│   ├── registry.go            # Tool registration (read-only & write modes)
│   ├── workbench_tools.go     # Workbench CRUD & status tools
│   ├── image_tools.go         # Image management tools
│   ├── hardware_profile_tools.go  # Hardware profile tools
│   ├── storage_tools.go       # PVC management tools
│   ├── namespace_tools.go     # Namespace listing tools
│   ├── pod_tools.go           # Pod listing tools
│   ├── resource_consumption_tools.go  # Resource monitoring tools
│   └── common.go              # Shared tool helpers
├── resources/                 # MCP resource definitions
├── prompts/                   # MCP prompts
└── Makefile                   # Build automation
```

### Adding New Tools

1. Implement the tool function in the appropriate file in `tools/`
2. Register it in `tools/registry.go`:
   - For read-only: Add to `RegisterReadOnlyTools()`
   - For write: Add to `RegisterWriteTools()`
3. Add tests in `tools/*_test.go`
4. Update this README

## Linting

This repository uses [golangci-lint](https://golangci-lint.run) for code quality checks.

### Install golangci-lint
```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### Run Linting

```bash
make lint
```


