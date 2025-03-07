# Tempural

A Go CLI tool for interacting with a Temporal server instance.

## Purpose

Tempural provides a simple command-line interface to interact with Temporal workflows. With Tempural, you can:

- List running workflows
- Start new workflows
- Get detailed information about a workflow
- Signal existing workflows
- Query workflow state

## Installation

### Using Go Install

You can install directly using Go:

```bash
go install github.com/weslien/tempural@latest
```

### Prerequisites

- Go 1.21 or higher
- Access to a Temporal server

### Building from source

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

If no workflow ID is provided, a random one will be generated.

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

## Examples

List all running workflows:
```bash
tempural list
```

Start a new workflow:
```bash
tempural start -t "ProcessOrder" -i '{"orderId": "12345"}'
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

## License

[MIT](LICENSE)