# Tempural

A Go CLI tool for interacting with a Temporal server instance.

## Purpose

Tempural provides a simple command-line interface to interact with Temporal workflows. With Tempural, you can:

- List running workflows
- Start new workflows
- Get detailed information about a workflow
- Signal existing workflows
- Query workflow state
- Infer parameter structure for workflows

## Installation

### Using Homebrew (macOS/Linux)

The easiest way to install tempural is via Homebrew:

```bash
# Add the tap repository
brew tap weslien/tap

# Install tempural
brew install tempural
```

Benefits of using Homebrew:
- Automatic dependency management
- Easy updates with `brew upgrade`
- No need to manually build from source
- Seamless integration with your system

To update to the latest version:

```bash
brew upgrade tempural
```

### Using Go Install

You can install directly using Go:

```bash
go install github.com/weslien/tempural@latest
```

### Prerequisites

- Go 1.21 or higher (only needed for Go installation method)
- Access to a Temporal server

### Building from source

Note: For most users, installing via Homebrew or Go Install is recommended instead of building from source.

```bash
# Clone the repository
git clone https://github.com/weslien/tempural.git
cd tempural

# Build using make
make build

# The binary will be in the build directory
./build/tempural -help

# Or install to your GOPATH/bin
make install
```

## Usage

Tempural provides several commands to interact with a Temporal server:

### Global Flags

These flags can be used with any command:

```
--address, -a     Temporal server address (default: "localhost:7233")
--namespace, -n   Temporal namespace (default: "default")
--task-queue, -q  Task queue for workflow execution (default: "default")
--workflow-id, -w Workflow ID for operations that require one
```

You can also set these values using environment variables:

```
TEMPORAL_ADDRESS
TEMPORAL_NAMESPACE
TEMPORAL_TASK_QUEUE
TEMPORAL_WORKFLOW_ID
```

### List Workflows

List all running workflows in the namespace:

```bash
tempural list
```

### Start a Workflow

Start a new workflow execution:

```bash
tempural start --workflow-type "YourWorkflowType" --input '{"key": "value"}'
```

Required flags:
- `--workflow-type, -t`: Type of workflow to start

Optional flags:
- `--input, -i`: JSON input for the workflow (default: "{}")
- `--workflow-id, -w, --id`: Explicit ID to use for the workflow
- `--interactive, --prompt`: Build workflow input interactively with prompts

If no workflow ID is provided (either via the command-specific `--workflow-id` flag or the global `-w` flag), a random one will be generated.

Examples:
```bash
# Start with auto-generated workflow ID
tempural start -t "ProcessOrder" -i '{"orderId": "12345"}'

# Start with specific workflow ID
tempural start -t "ProcessOrder" -i '{"orderId": "12345"}' --workflow-id "order-12345"

# Using the global workflow ID flag
tempural -w "order-12345" start -t "ProcessOrder" -i '{"orderId": "12345"}'
```

#### Interactive Mode

The interactive mode provides a guided experience for building workflow input:

```bash
tempural start -t "YourWorkflowType" --interactive
```

When using interactive mode:
1. The CLI first tries to infer the expected parameter structure from existing workflows of the same type
2. If it finds a matching schema, it guides you through filling in each field with the correct type
3. If no schema is found, it provides a generic input builder that can create any JSON structure
4. After building the input, it shows you the final JSON and asks for confirmation before starting the workflow

This is especially useful when you're not familiar with the exact structure of parameters a workflow expects.

### Describe a Workflow

Get detailed information about a specific workflow:

```bash
tempural describe --workflow-id "your-workflow-id"
```

Required flags:
- `--workflow-id, -w`: ID of the workflow to describe (can be provided globally)

Optional flags:
- `--run-id, -r`: Run ID of the workflow (if not provided, the latest run will be used)

This command shows detailed information about the workflow, including:
- Basic workflow metadata (ID, type, status)
- Timing information (start time, execution time)
- Input data the workflow was started with
- Pending activities (if any)
- Pending child workflows (if any)

Both command formats are supported:
```bash
# These are equivalent:
tempural --workflow-id "your-workflow-id" describe
tempural describe --workflow-id "your-workflow-id"
```

### Signal a Workflow

Send a signal to a running workflow:

```bash
tempural signal --workflow-id "your-workflow-id" --signal-name "YourSignalName" --input '{"key": "value"}'
```

Required flags:
- `--signal-name, -s`: Name of the signal to send
- `--workflow-id, -w`: ID of the workflow to signal (can be provided globally)

Optional flags:
- `--input, -i`: JSON input for the signal (default: "{}")

### Query a Workflow

Query the state of a running workflow:

```bash
tempural query --workflow-id "your-workflow-id" --query-type "YourQueryType" --args '{"key": "value"}'
```

Required flags:
- `--query-type, -q`: Type of query to execute
- `--workflow-id, -w`: ID of the workflow to query (can be provided globally)

Optional flags:
- `--args, -a`: JSON arguments for the query (default: "{}")

### Infer Workflow Parameters

Discover the parameter structure a workflow type expects by examining past executions:

```bash
tempural infer-params --workflow-type "YourWorkflowType"
```

Required flags:
- `--workflow-type, -t`: Type of workflow to infer parameters for

Optional flags:
- `--limit, -l`: Maximum number of workflows to examine (default: 3)
- `--json-schema, -j`: Output as JSONSchema format instead of examples
- `--raw, -r`: Output raw schema without pretty-printing (useful for piping to files)

This command:
1. Finds recent executions of the specified workflow type
2. Examines their input parameters
3. Shows example parameter structures that can be used as templates

With the `--json-schema` flag, it generates a formal JSONSchema representation that can be used for validation, documentation, or code generation.

Example JSONSchema output:
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "title": "ProcessOrder Parameters",
  "description": "Parameter schema for ProcessOrder workflow",
  "properties": {
    "orderId": {
      "type": "string"
    },
    "customerId": {
      "type": "string"
    },
    "items": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "productId": {
            "type": "string"
          },
          "quantity": {
            "type": "number"
          }
        },
        "required": ["productId", "quantity"]
      }
    }
  },
  "required": ["orderId", "customerId", "items"]
}
```

This is particularly useful when you're unsure about the structure of parameters a workflow expects.

## Examples

List all running workflows:
```bash
tempural list
```

Start a new workflow:
```bash
# Auto-generated workflow ID
tempural start -t "ProcessOrder" -i '{"orderId": "12345"}'

# Specific workflow ID
tempural start -t "ProcessOrder" -i '{"orderId": "12345"}' --workflow-id "order-12345"
```

Start a workflow with interactive input:
```bash
tempural start -t "ProcessOrder" --interactive
```

Get detailed information about a workflow:
```bash
tempural describe -w "workflow-1234567890"
```

Signal a workflow:
```bash
tempural signal -w "workflow-1234567890" -s "CancelOrder" -i '{"reason": "Customer request"}'
```

Query a workflow:
```bash
tempural query -w "workflow-1234567890" -q "GetOrderStatus"
```

Infer parameters for a workflow type:
```bash
tempural infer-params -t "ProcessOrder"
```

Generate JSONSchema for a workflow type:
```bash
# Output JSONSchema to console with pretty formatting
tempural infer-params -t "ProcessOrder" --json-schema

# Save JSONSchema to a file
tempural infer-params -t "ProcessOrder" --json-schema --raw > process-order-schema.json
```

## License

[MIT](LICENSE)