package app

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/urfave/cli/v2"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/history/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
)

// TemporalConfig holds configuration for connecting to Temporal
type TemporalConfig struct {
	Address    string
	Namespace  string
	TaskQueue  string
	WorkflowID string
}

// NewTemporalCLI creates a new CLI application for interacting with Temporal
func NewTemporalCLI() *cli.App {
	var config TemporalConfig

	app := &cli.App{
		Name:                   "tempural",
		Usage:                  "A CLI tool for interacting with Temporal server",
		UseShortOptionHandling: true,
		EnableBashCompletion:   true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "address",
				Aliases:     []string{"a"},
				Value:       "localhost:7233",
				Usage:       "Temporal server address",
				Destination: &config.Address,
				EnvVars:     []string{"TEMPORAL_ADDRESS"},
			},
			&cli.StringFlag{
				Name:        "namespace",
				Aliases:     []string{"n"},
				Value:       "default",
				Usage:       "Temporal namespace",
				Destination: &config.Namespace,
				EnvVars:     []string{"TEMPORAL_NAMESPACE"},
			},
			&cli.StringFlag{
				Name:        "task-queue",
				Aliases:     []string{"q"},
				Value:       "default",
				Usage:       "Task queue for workflow execution",
				Destination: &config.TaskQueue,
				EnvVars:     []string{"TEMPORAL_TASK_QUEUE"},
			},
			&cli.StringFlag{
				Name:        "workflow-id",
				Aliases:     []string{"w"},
				Value:       "",
				Usage:       "Workflow ID for operations that require one",
				Destination: &config.WorkflowID,
				EnvVars:     []string{"TEMPORAL_WORKFLOW_ID"},
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "list",
				Usage: "List running workflows",
				Action: func(c *cli.Context) error {
					return listWorkflows(c, config)
				},
			},
			{
				Name:  "start",
				Usage: "Start a new workflow",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "workflow-type",
						Aliases:  []string{"t"},
						Usage:    "Type of workflow to start",
						Value:    "",
						Required: true,
					},
					&cli.StringFlag{
						Name:    "input",
						Aliases: []string{"i"},
						Usage:   "JSON input for the workflow",
						Value:   "{}",
					},
				},
				Action: func(c *cli.Context) error {
					return startWorkflow(c, config)
				},
			},
			{
				Name:  "describe",
				Usage: "Get detailed information about a workflow",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "run-id",
						Aliases: []string{"r"},
						Usage:   "Run ID of the workflow (optional)",
						Value:   "",
					},
				},
				Action: func(c *cli.Context) error {
					return describeWorkflow(c, config)
				},
			},
			{
				Name:  "signal",
				Usage: "Signal a workflow",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "signal-name",
						Aliases:  []string{"s"},
						Usage:    "Name of the signal to send",
						Required: true,
					},
					&cli.StringFlag{
						Name:    "input",
						Aliases: []string{"i"},
						Usage:   "JSON input for the signal",
						Value:   "{}",
					},
				},
				Action: func(c *cli.Context) error {
					return signalWorkflow(c, config)
				},
			},
			{
				Name:  "query",
				Usage: "Query a workflow",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "query-type",
						Aliases:  []string{"q"},
						Usage:    "Type of query to execute",
						Required: true,
					},
					&cli.StringFlag{
						Name:    "args",
						Aliases: []string{"a"},
						Usage:   "JSON arguments for the query",
						Value:   "{}",
					},
				},
				Action: func(c *cli.Context) error {
					return queryWorkflow(c, config)
				},
			},
		},
	}

	return app
}

// getTemporalClient creates a new Temporal client
func getTemporalClient(config TemporalConfig) (client.Client, error) {
	return client.Dial(client.Options{
		HostPort:  config.Address,
		Namespace: config.Namespace,
	})
}

// listWorkflows lists running workflows
func listWorkflows(c *cli.Context, config TemporalConfig) error {
	temporalClient, err := getTemporalClient(config)
	if err != nil {
		return fmt.Errorf("failed to create Temporal client: %w", err)
	}
	defer temporalClient.Close()

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Use the native API to list workflows
	listRequest := &workflowservice.ListWorkflowExecutionsRequest{
		// Default to open workflows
		Query: fmt.Sprintf("ExecutionStatus=%d", int32(enums.WORKFLOW_EXECUTION_STATUS_RUNNING)),
	}

	resp, err := temporalClient.ListWorkflow(ctx, listRequest)
	if err != nil {
		return fmt.Errorf("failed to list workflows: %w", err)
	}

	// Display the results
	fmt.Printf("Found %d workflows:\n", len(resp.Executions))
	for i, execution := range resp.Executions {
		fmt.Printf("%d. ID: %s, Type: %s, Status: %s\n",
			i+1,
			execution.Execution.WorkflowId,
			execution.Type.Name,
			enums.WorkflowExecutionStatus_name[int32(execution.Status)],
		)
	}

	return nil
}

// startWorkflow starts a new workflow
func startWorkflow(c *cli.Context, config TemporalConfig) error {
	temporalClient, err := getTemporalClient(config)
	if err != nil {
		return fmt.Errorf("failed to create Temporal client: %w", err)
	}
	defer temporalClient.Close()

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	workflowType := c.String("workflow-type")
	input := c.String("input")
	workflowID := config.WorkflowID

	// Generate a random workflow ID if not provided
	if workflowID == "" {
		workflowID = fmt.Sprintf("workflow-%d", time.Now().Unix())
	}

	// Start the workflow
	workflowOptions := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: config.TaskQueue,
	}

	// We're using an empty payloads here, but in a real implementation
	// you would parse the input JSON and convert it to the appropriate type
	we, err := temporalClient.ExecuteWorkflow(ctx, workflowOptions, workflowType, []byte(input))
	if err != nil {
		return fmt.Errorf("failed to start workflow: %w", err)
	}

	fmt.Printf("Started workflow execution\n")
	fmt.Printf("Workflow ID: %s\n", we.GetID())
	fmt.Printf("Run ID: %s\n", we.GetRunID())

	return nil
}

// signalWorkflow signals a workflow
func signalWorkflow(c *cli.Context, config TemporalConfig) error {
	if config.WorkflowID == "" {
		return fmt.Errorf("workflow ID is required for signaling")
	}

	temporalClient, err := getTemporalClient(config)
	if err != nil {
		return fmt.Errorf("failed to create Temporal client: %w", err)
	}
	defer temporalClient.Close()

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	signalName := c.String("signal-name")
	input := c.String("input")

	// Signal the workflow
	err = temporalClient.SignalWorkflow(ctx, config.WorkflowID, "", signalName, []byte(input))
	if err != nil {
		return fmt.Errorf("failed to signal workflow: %w", err)
	}

	fmt.Printf("Signal '%s' sent to workflow ID: %s\n", signalName, config.WorkflowID)
	return nil
}

// queryWorkflow queries a workflow
func queryWorkflow(c *cli.Context, config TemporalConfig) error {
	if config.WorkflowID == "" {
		return fmt.Errorf("workflow ID is required for querying")
	}

	temporalClient, err := getTemporalClient(config)
	if err != nil {
		return fmt.Errorf("failed to create Temporal client: %w", err)
	}
	defer temporalClient.Close()

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	queryType := c.String("query-type")
	args := c.String("args")

	// Query the workflow
	response, err := temporalClient.QueryWorkflow(ctx, config.WorkflowID, "", queryType, []byte(args))
	if err != nil {
		return fmt.Errorf("failed to query workflow: %w", err)
	}

	// Get the result
	var result interface{}
	if err := response.Get(&result); err != nil {
		return fmt.Errorf("failed to parse query result: %w", err)
	}

	fmt.Printf("Query result: %v\n", result)
	return nil
}

// describeWorkflow gets detailed information about a specific workflow
func describeWorkflow(c *cli.Context, config TemporalConfig) error {
	if config.WorkflowID == "" {
		return fmt.Errorf("workflow ID is required")
	}

	temporalClient, err := getTemporalClient(config)
	if err != nil {
		return fmt.Errorf("failed to create Temporal client: %w", err)
	}
	defer temporalClient.Close()

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	runID := c.String("run-id")

	// Get workflow execution details
	resp, err := temporalClient.DescribeWorkflowExecution(ctx, config.WorkflowID, runID)
	if err != nil {
		return fmt.Errorf("failed to describe workflow: %w", err)
	}

	// Print execution details
	fmt.Println(fmt.Sprintf("%s%s==== Workflow Details ====%s", colorBold, colorBlue, colorReset))
	execution := resp.WorkflowExecutionInfo

	fmt.Printf("Workflow ID: %s\n", execution.Execution.WorkflowId)
	fmt.Printf("Run ID: %s\n", execution.Execution.RunId)
	fmt.Printf("Type: %s\n", execution.Type.Name)
	fmt.Printf("Status: %s\n", enums.WorkflowExecutionStatus_name[int32(execution.Status)])

	// Time values are displayed as ISO
	fmt.Printf("Start Time: %v\n", execution.StartTime)

	if execution.CloseTime != nil {
		fmt.Printf("Close Time: %v\n", execution.CloseTime)
	}

	fmt.Printf("History Length: %d\n", execution.HistoryLength)
	fmt.Printf("Execution Time: %v\n", execution.ExecutionTime)

	// Fetch workflow history to get input data
	iter := temporalClient.GetWorkflowHistory(ctx, config.WorkflowID, runID, false, 0)
	var startedEvent *history.HistoryEvent
	var startedEventFound bool

	for iter.HasNext() {
		event, err := iter.Next()
		if err != nil {
			return fmt.Errorf("failed to get workflow history: %w", err)
		}

		// Look for the workflow started event (always the first event)
		if event.GetEventType() == enums.EVENT_TYPE_WORKFLOW_EXECUTION_STARTED {
			startedEvent = event
			startedEventFound = true
			break
		}
	}

	// Display input if found
	if startedEventFound && startedEvent != nil {
		startedAttrs := startedEvent.GetWorkflowExecutionStartedEventAttributes()
		if startedAttrs != nil && startedAttrs.Input != nil {
			fmt.Println(fmt.Sprintf("\n%s%s==== Workflow Input ====%s", colorBold, colorGreen, colorReset))

			// Get the payloads from the input
			payloads := startedAttrs.Input.GetPayloads()
			if len(payloads) > 0 {
				for i, payload := range payloads {
					fmt.Printf("Input %d:\n", i+1)

					// Try to extract data as string and parse as JSON
					data := payload.GetData()
					if len(data) > 0 {
						// Try to parse as JSON
						var jsonObj interface{}
						if err := json.Unmarshal(data, &jsonObj); err == nil {
							// If successful, pretty print the JSON
							prettyJSON, err := json.MarshalIndent(jsonObj, "  ", "  ")
							if err == nil {
								fmt.Printf("  %s\n", string(prettyJSON))
							} else {
								fmt.Printf("  %s\n", string(data))
							}
						} else {
							// If not JSON, print as string
							fmt.Printf("  %s\n", string(data))
						}
					} else {
						fmt.Println("  <empty>")
					}
				}
			} else {
				fmt.Println("No input data provided")
			}
		}
	}

	// Print more workflow details
	fmt.Println(fmt.Sprintf("\n%s%s==== Pending Activities ====%s", colorBold, colorMagenta, colorReset))
	if len(resp.PendingActivities) == 0 {
		fmt.Println("No pending activities")
	} else {
		for i, activity := range resp.PendingActivities {
			fmt.Printf("Activity %d:\n", i+1)
			fmt.Printf("  Type: %s\n", activity.ActivityType.Name)
			fmt.Printf("  State: %s\n", activity.State.String())
			fmt.Printf("  Scheduled Time: %v\n", activity.ScheduledTime)
			if activity.LastHeartbeatTime != nil {
				fmt.Printf("  Last Heartbeat: %v\n", activity.LastHeartbeatTime)
			}
		}
	}

	// Print pending children workflows if any
	if len(resp.PendingChildren) > 0 {
		fmt.Println(fmt.Sprintf("\n%s%s==== Pending Child Workflows ====%s", colorBold, colorCyan, colorReset))
		for i, child := range resp.PendingChildren {
			fmt.Printf("Child Workflow %d:\n", i+1)
			fmt.Printf("  ID: %s\n", child.WorkflowId)
			fmt.Printf("  Type: %s\n", child.WorkflowTypeName)
			fmt.Printf("  Run ID: %s\n", child.RunId)
			fmt.Printf("  Parent Close Policy: %s\n", child.ParentClosePolicy.String())
		}
	}

	return nil
}
