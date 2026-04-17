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

Run tests using Make or directly with Go:

```bash
# Run all tests with Make
make test

# Run tests directly with verbose output
go test -v ./tools/... ./resources/... ./prompts/...
```

Test coverage includes:
- Tool operations (workbenches, images, hardware profiles)
- Resource management
- Image catalog parsing

## Evaluation

This project uses [Promptfoo](https://www.promptfoo.dev/) for automated evaluation of tool selection and execution quality. Promptfoo is an open-source testing framework specifically designed for evaluating LLM applications, tool calling, and agent behavior. It provides metrics-based testing with visual reports.

### Run Evaluation

```bash
# Run evaluation tests
make eval

# Run evaluation and view results in browser
make eval-view
```

**API Key Requirements:**

The evaluation uses AI models to test the MCP server. You'll need an API key for one of the supported providers:

- **OpenAI** (default): Set `OPENAI_API_KEY` environment variable
  ```bash
  export OPENAI_API_KEY=your-key-here
  make eval
  ```

- **Anthropic Claude**: Uncomment the Claude provider in `promptfoo.yaml` and set `ANTHROPIC_API_KEY`
  ```bash
  export ANTHROPIC_API_KEY=your-key-here
  make eval
  ```

- **Google Gemini**: Uncomment the Gemini provider in `promptfoo.yaml` and set `GOOGLE_API_KEY`
  ```bash
  export GOOGLE_API_KEY=your-key-here
  make eval
  ```

The evaluation measures:
- **Tool Selection Accuracy** - Does the system choose the correct tool for each prompt?
- **Parameter Extraction** - Are the tool parameters extracted correctly?
- **Execution Success Rate** - Do the tools execute without errors?

Results are displayed in an interactive web UI showing pass/fail rates, detailed comparisons, and performance metrics.

**Prerequisites**: Node.js (^20.20.0 or >=22.22.0) and npm must be installed.

### Evaluation Findings

Evaluation was run against Google Gemini models (2.5 Flash and 2.5 Pro). Key findings:

| Test Case | Gemini 2.5 Flash | Gemini 2.5 Pro |
|-----------|-----------------|----------------|
| List workbenches in a namespace | PASS | PASS |
| Resource consumption per namespace | PASS | PASS |
| List all workbenches across namespaces | FAIL | FAIL |
| Create a workbench | FAIL | FAIL |

**Observations:**

1. **`list_all_workbenches` tool confusion** — Both models refused to call this tool despite it being available. The tool's input schema requires a `namespace` parameter, which contradicts its description ("list workbenches across all namespaces"). Models interpret this inconsistency as a reason to fall back to the per-namespace `list_workbenches` tool or refuse entirely.

2. **`create_workbench` tool refusal** — Both models declined to call `create_workbench`, claiming they could only list resources. The tool has a complex input schema with many required fields (`hardwareProfile` as a nested object, `imageTag`, `pvcName`, etc.). When the test prompt does not supply all required parameters, the models choose not to attempt the call rather than inferring or requesting missing values.

3. **Model quality vs. cost trade-off** — Gemini 2.5 Flash and 2.5 Pro produced identical results for tool selection tasks. Lighter models may struggle with complex tool schemas. Note that `gemini-2.0-flash` has been deprecated for new API users as of April 2026.

4. **Free tier limitations** — The Google Gemini free tier quota can be exhausted quickly. When the quota is hit (`limit: 0`), promptfoo enters an infinite retry loop. A pay-as-you-go billing plan is recommended for running evaluations reliably.

**Recommendations for improving eval pass rates:**
- Remove the contradictory `namespace` parameter from the `list_all_workbenches` input schema, or rename the tool to clarify its behavior.
- Enrich test prompts for write operations to include all required parameters (e.g., hardware profile, image tag, PVC name), or simplify the tool's required fields so the model can attempt the call with partial information.
- Consider testing with OpenAI GPT-4o or Anthropic Claude, which may handle complex tool schemas more reliably.

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

The mode is controlled by the `MCP_RHOAI_MODE` environment variable:

| Variable Value | Mode | Description |
|---------------|------|-------------|
| Not set or any value except `write` | Read-only | Default, safe mode |
| `write` | Read-write | Full access to all tools |

Set the environment variable in your MCP server configuration (see [Configuration](#configuration) section above).

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


