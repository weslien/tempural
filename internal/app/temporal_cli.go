package app

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/history/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
)

// TemporalConfig holds configuration for connecting to Temporal
type TemporalConfig struct {
	Address         string
	Namespace       string
	TaskQueue       string
	WorkflowID      string
	Debug           bool
	CPUProfile      string
	MemProfile      string
	EnableProfiling bool
	ProfilePort     int
}

// NewTemporalCLI creates a new CLI application for interacting with Temporal
func NewTemporalCLI() *cli.App {
	var config TemporalConfig

	app := &cli.App{
		Name:                   "tempural",
		Usage:                  "A CLI tool for interacting with Temporal server",
		UseShortOptionHandling: true,
		EnableBashCompletion:   true,
		AllowExtFlags:          true, // Allow flags after commands
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
			&cli.BoolFlag{
				Name:        "debug",
				Aliases:     []string{"d"},
				Usage:       "Enable debug mode with verbose logging",
				Destination: &config.Debug,
				Value:       false,
			},
			&cli.StringFlag{
				Name:        "cpu-profile",
				Usage:       "Write CPU profile to file",
				Destination: &config.CPUProfile,
				Value:       "",
			},
			&cli.StringFlag{
				Name:        "mem-profile",
				Usage:       "Write memory profile to file",
				Destination: &config.MemProfile,
				Value:       "",
			},
			&cli.BoolFlag{
				Name:        "pprof",
				Usage:       "Enable runtime profiling server",
				Destination: &config.EnableProfiling,
				Value:       false,
			},
			&cli.IntFlag{
				Name:        "pprof-port",
				Usage:       "Port for runtime profiling server",
				Destination: &config.ProfilePort,
				Value:       6060,
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
					&cli.StringFlag{
						Name:    "workflow-id",
						Aliases: []string{"w", "id"},
						Usage:   "Specific ID to use for the workflow (overrides global workflow-id flag)",
						Value:   "",
					},
					&cli.BoolFlag{
						Name:    "interactive",
						Aliases: []string{"prompt"},
						Usage:   "Build workflow input interactively with prompts",
						Value:   false,
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
			{
				Name:  "infer-params",
				Usage: "Infer parameter structure for a workflow type",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "workflow-type",
						Aliases:  []string{"t"},
						Usage:    "Type of workflow to infer parameters for",
						Required: true,
					},
					&cli.IntFlag{
						Name:    "limit",
						Aliases: []string{"l"},
						Usage:   "Maximum number of workflows to examine",
						Value:   3,
					},
					&cli.BoolFlag{
						Name:    "json-schema",
						Aliases: []string{"j"},
						Usage:   "Output as JSONSchema",
						Value:   false,
					},
					&cli.BoolFlag{
						Name:    "raw",
						Aliases: []string{"r"},
						Usage:   "Output raw schema without pretty printing",
						Value:   false,
					},
				},
				Action: func(c *cli.Context) error {
					return inferWorkflowParams(c, config)
				},
			},
		},
		Before: func(c *cli.Context) error {
			// Setup profiling before any command runs
			return SetupProfiling(config)
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

	// First check for a command-specific workflow ID flag
	workflowID := c.String("workflow-id")

	// If not provided, fall back to the global flag
	if workflowID == "" {
		workflowID = config.WorkflowID
	}

	// Generate a random workflow ID if still not provided
	if workflowID == "" {
		workflowID = fmt.Sprintf("workflow-%d", time.Now().Unix())
		fmt.Printf("No workflow ID provided, using auto-generated ID: %s\n", workflowID)
	}

	var input string
	var inputData interface{}

	// Handle interactive mode if enabled
	if c.Bool("interactive") {
		fmt.Printf("%sInteractive Mode: Build input for workflow %s%s%s\n",
			colorBold, colorBlue, workflowType, colorReset)

		// Try to infer workflow parameters if available
		schema, err := inferWorkflowSchemaForType(ctx, temporalClient, workflowType)
		if err != nil {
			fmt.Printf("%sNote:%s Couldn't find existing workflows to infer parameters, using generic input.\n\n",
				colorYellow, colorReset)
			// Build generic input when no schema is available
			inputData = buildInputInteractively(nil)
		} else {
			// Build input based on schema
			inputData = buildInputInteractivelyFromSchema(schema)
		}

		// Convert the input data to JSON
		jsonBytes, err := json.Marshal(inputData)
		if err != nil {
			return fmt.Errorf("failed to marshal input data: %w", err)
		}
		input = string(jsonBytes)

		// Show the final input
		fmt.Printf("\n%sFinal Input:%s\n", colorGreen, colorReset)
		prettyJSON, _ := json.MarshalIndent(inputData, "", "  ")
		fmt.Println(string(prettyJSON))
		fmt.Println()

		// Confirm with user
		if !confirmAction("Start workflow with this input?") {
			return fmt.Errorf("workflow start canceled by user")
		}
	} else {
		// Get input from flag value
		inputFlag := c.String("input")

		// Check if input should be read from stdin
		if inputFlag == "-" {
			fmt.Println("Reading input from stdin...")
			scanner := bufio.NewScanner(os.Stdin)
			var inputBuilder strings.Builder
			for scanner.Scan() {
				inputBuilder.WriteString(scanner.Text())
			}
			if err := scanner.Err(); err != nil {
				return fmt.Errorf("error reading from stdin: %w", err)
			}
			input = inputBuilder.String()

			// If input is empty, provide a warning
			if strings.TrimSpace(input) == "" {
				fmt.Println("Warning: Empty input received from stdin")
				input = "{}" // Fallback to empty JSON object
			}
		} else {
			// Use the input provided in the flag
			input = inputFlag
		}
	}

	// Start the workflow
	workflowOptions := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: config.TaskQueue,
	}

	we, err := temporalClient.ExecuteWorkflow(ctx, workflowOptions, workflowType, []byte(input))
	if err != nil {
		return fmt.Errorf("failed to start workflow: %w", err)
	}

	fmt.Printf("Started workflow execution\n")
	fmt.Printf("Workflow ID: %s\n", we.GetID())
	fmt.Printf("Run ID: %s\n", we.GetRunID())

	return nil
}

// inferWorkflowSchemaForType attempts to infer a schema for a workflow type
// by examining recent executions
func inferWorkflowSchemaForType(ctx context.Context, client client.Client, workflowType string) (map[string]interface{}, error) {
	// Find recent workflows of this type
	query := fmt.Sprintf("WorkflowType='%s'", workflowType)
	listRequest := &workflowservice.ListWorkflowExecutionsRequest{
		Query: query,
	}

	resp, err := client.ListWorkflow(ctx, listRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to list workflows: %w", err)
	}

	if len(resp.Executions) == 0 {
		return nil, fmt.Errorf("no workflows of type '%s' found", workflowType)
	}

	// Examine the first workflow we find to extract its parameter structure
	execution := resp.Executions[0]
	workflowID := execution.Execution.WorkflowId
	runID := execution.Execution.RunId

	// Get workflow history to find start event
	iter := client.GetWorkflowHistory(ctx, workflowID, runID, false, 0)

	// Look for first event (workflow started)
	for iter.HasNext() {
		event, err := iter.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to fetch history: %w", err)
		}

		if event.GetEventType() == enums.EVENT_TYPE_WORKFLOW_EXECUTION_STARTED {
			startedAttrs := event.GetWorkflowExecutionStartedEventAttributes()
			if startedAttrs != nil && startedAttrs.Input != nil {
				payloads := startedAttrs.Input.GetPayloads()
				if len(payloads) > 0 {
					for _, payload := range payloads {
						data := payload.GetData()
						if len(data) > 0 {
							// Try to parse as JSON
							var jsonObj interface{}
							if err := json.Unmarshal(data, &jsonObj); err == nil {
								// Convert to schema representation
								return generateJSONSchema(jsonObj, ""), nil
							}
						}
					}
				}
			}
			break // Only check the first event
		}
	}

	return nil, fmt.Errorf("no valid input found in workflow history")
}

// buildInputInteractivelyFromSchema builds a workflow input object interactively based on a schema
func buildInputInteractivelyFromSchema(schema map[string]interface{}) interface{} {
	if schema == nil {
		return buildInputInteractively(nil)
	}

	fmt.Printf("Building input based on inferred schema:\n\n")

	// Check if we have a properties field (indicates an object)
	if properties, ok := schema["properties"].(map[string]interface{}); ok {
		// Build an object with the properties from the schema
		result := make(map[string]interface{})

		// Get required fields if any
		required := make(map[string]bool)
		if reqFields, ok := schema["required"].([]string); ok {
			for _, field := range reqFields {
				required[field] = true
			}
		}

		for fieldName, fieldSchema := range properties {
			fieldSchemaObj, ok := fieldSchema.(map[string]interface{})
			if !ok {
				continue
			}

			fieldType, _ := fieldSchemaObj["type"].(string)
			isRequired := required[fieldName]

			// Prompt for this field
			fmt.Printf("Field: %s%s%s", colorBold, fieldName, colorReset)
			if isRequired {
				fmt.Printf(" %s(required)%s", colorRed, colorReset)
			}
			fmt.Printf(" [%s]\n", fieldType)

			// Handle field based on its type
			var value interface{}

			switch fieldType {
			case "object":
				// Recursively build nested object
				fmt.Printf("Enter values for the nested object '%s':\n", fieldName)
				value = buildInputInteractivelyFromSchema(fieldSchemaObj)

			case "array":
				// Build array
				items, ok := fieldSchemaObj["items"].(map[string]interface{})
				if !ok {
					items = map[string]interface{}{"type": "string"}
				}

				value = buildArrayInteractively(items)

			case "string":
				value = promptForInput(fmt.Sprintf("Enter value for %s", fieldName), isRequired)

			case "number", "integer":
				for {
					input := promptForInput(fmt.Sprintf("Enter number for %s", fieldName), isRequired)
					if input == "" && !isRequired {
						break
					}

					if num, err := strconv.ParseFloat(input, 64); err == nil {
						value = num
						break
					} else {
						fmt.Printf("%sError:%s Please enter a valid number\n", colorRed, colorReset)
					}
				}

			case "boolean":
				for {
					input := promptForInput(fmt.Sprintf("Enter true/false for %s", fieldName), isRequired)
					if input == "" && !isRequired {
						break
					}

					lowerInput := strings.ToLower(input)
					if lowerInput == "true" || lowerInput == "yes" || lowerInput == "y" {
						value = true
						break
					} else if lowerInput == "false" || lowerInput == "no" || lowerInput == "n" {
						value = false
						break
					} else {
						fmt.Printf("%sError:%s Please enter true or false\n", colorRed, colorReset)
					}
				}

			default:
				// Default to string for unknown types
				value = promptForInput(fmt.Sprintf("Enter value for %s", fieldName), isRequired)
			}

			// Add to result if a value was provided
			if value != nil && value != "" {
				result[fieldName] = value
			}
		}

		return result
	}

	// If we don't have properties, fall back to generic input
	return buildInputInteractively(nil)
}

// buildArrayInteractively prompts the user to build an array item by item
func buildArrayInteractively(itemSchema map[string]interface{}) []interface{} {
	result := make([]interface{}, 0)

	fmt.Printf("Building array (enter empty value when done):\n")

	itemType, _ := itemSchema["type"].(string)

	for i := 1; ; i++ {
		fmt.Printf("Item %d:\n", i)

		var item interface{}
		var done bool

		switch itemType {
		case "object":
			// Build an object for this array item
			item = buildInputInteractivelyFromSchema(itemSchema)
			// Check if the object is empty
			if objItem, ok := item.(map[string]interface{}); ok && len(objItem) == 0 {
				done = true
			}

		case "array":
			// Handle nested arrays (rare but possible)
			nestedItems, ok := itemSchema["items"].(map[string]interface{})
			if !ok {
				nestedItems = map[string]interface{}{"type": "string"}
			}

			nestedArray := buildArrayInteractively(nestedItems)
			if len(nestedArray) == 0 {
				done = true
			} else {
				item = nestedArray
			}

		case "string":
			input := promptForInput("Enter value (or empty to finish)", false)
			if input == "" {
				done = true
			} else {
				item = input
			}

		case "number", "integer":
			input := promptForInput("Enter number (or empty to finish)", false)
			if input == "" {
				done = true
			} else {
				if num, err := strconv.ParseFloat(input, 64); err == nil {
					item = num
				} else {
					fmt.Printf("%sError:%s Not a valid number, skipping\n", colorRed, colorReset)
					continue
				}
			}

		case "boolean":
			input := promptForInput("Enter true/false (or empty to finish)", false)
			if input == "" {
				done = true
			} else {
				lowerInput := strings.ToLower(input)
				if lowerInput == "true" || lowerInput == "yes" || lowerInput == "y" {
					item = true
					break
				} else if lowerInput == "false" || lowerInput == "no" || lowerInput == "n" {
					item = false
					break
				}
				fmt.Printf("%sError:%s Please enter true or false\n", colorRed, colorReset)
			}

		default:
			// Default to string
			input := promptForInput("Enter value (or empty to finish)", false)
			if input == "" {
				done = true
			} else {
				item = input
			}
		}

		if done {
			break
		}

		result = append(result, item)
	}

	return result
}

// buildInputInteractively guides the user through building a JSON object from scratch
func buildInputInteractively(parent interface{}) interface{} {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("Select the type of input to create:")
	fmt.Println("1. Object (JSON object with key/value pairs)")
	fmt.Println("2. Array (List of values)")
	fmt.Println("3. String")
	fmt.Println("4. Number")
	fmt.Println("5. Boolean (true/false)")
	fmt.Println("6. Empty object ({})")
	fmt.Println("7. Empty array ([])")
	fmt.Println("8. null")

	for {
		fmt.Print("Enter your choice (1-8): ")
		scanner.Scan()
		choice := strings.TrimSpace(scanner.Text())

		switch choice {
		case "1":
			return buildObjectInteractively(scanner)
		case "2":
			return buildArrayInteractivelyGeneric(scanner)
		case "3":
			fmt.Print("Enter string value: ")
			scanner.Scan()
			return scanner.Text()
		case "4":
			for {
				fmt.Print("Enter number value: ")
				scanner.Scan()
				numStr := strings.TrimSpace(scanner.Text())
				if num, err := strconv.ParseFloat(numStr, 64); err == nil {
					return num
				}
				fmt.Printf("%sError:%s Please enter a valid number\n", colorRed, colorReset)
			}
		case "5":
			for {
				fmt.Print("Enter boolean value (true/false): ")
				scanner.Scan()
				boolStr := strings.ToLower(strings.TrimSpace(scanner.Text()))
				if boolStr == "true" || boolStr == "yes" || boolStr == "y" {
					return true
				} else if boolStr == "false" || boolStr == "no" || boolStr == "n" {
					return false
				}
				fmt.Printf("%sError:%s Please enter true or false\n", colorRed, colorReset)
			}
		case "6":
			return map[string]interface{}{}
		case "7":
			return []interface{}{}
		case "8":
			return nil
		default:
			fmt.Printf("%sError:%s Please enter a number between 1 and 8\n", colorRed, colorReset)
		}
	}
}

// buildObjectInteractively prompts the user to build a JSON object
func buildObjectInteractively(scanner *bufio.Scanner) map[string]interface{} {
	result := make(map[string]interface{})

	fmt.Println("\nBuilding an object. Enter empty key to finish.")

	for {
		fmt.Print("Enter field name (or empty to finish): ")
		scanner.Scan()
		key := strings.TrimSpace(scanner.Text())

		if key == "" {
			break
		}

		// Recursive call to build the field value
		fmt.Printf("Set value for '%s':\n", key)
		value := buildInputInteractively(result)

		// Add to result
		result[key] = value
	}

	return result
}

// buildArrayInteractivelyGeneric prompts the user to build a generic array
func buildArrayInteractivelyGeneric(scanner *bufio.Scanner) []interface{} {
	result := make([]interface{}, 0)

	fmt.Println("\nBuilding an array. Enter empty value to finish.")

	for i := 1; ; i++ {
		fmt.Printf("Item %d:\n", i)

		fmt.Println("Select the type of the array item:")
		fmt.Println("1. Object (JSON object with key/value pairs)")
		fmt.Println("2. Array (List of values)")
		fmt.Println("3. String")
		fmt.Println("4. Number")
		fmt.Println("5. Boolean (true/false)")
		fmt.Println("6. null")
		fmt.Println("7. Done (finish array)")

		fmt.Print("Enter your choice (1-7): ")
		scanner.Scan()
		choice := strings.TrimSpace(scanner.Text())

		if choice == "7" {
			break
		}

		var value interface{}

		switch choice {
		case "1":
			value = buildObjectInteractively(scanner)
		case "2":
			value = buildArrayInteractivelyGeneric(scanner)
		case "3":
			fmt.Print("Enter string value: ")
			scanner.Scan()
			value = scanner.Text()
		case "4":
			for {
				fmt.Print("Enter number value: ")
				scanner.Scan()
				numStr := strings.TrimSpace(scanner.Text())
				if num, err := strconv.ParseFloat(numStr, 64); err == nil {
					value = num
					break
				}
				fmt.Printf("%sError:%s Please enter a valid number\n", colorRed, colorReset)
			}
		case "5":
			for {
				fmt.Print("Enter boolean value (true/false): ")
				scanner.Scan()
				boolStr := strings.ToLower(strings.TrimSpace(scanner.Text()))
				if boolStr == "true" || boolStr == "yes" || boolStr == "y" {
					value = true
					break
				} else if boolStr == "false" || boolStr == "no" || boolStr == "n" {
					value = false
					break
				}
				fmt.Printf("%sError:%s Please enter true or false\n", colorRed, colorReset)
			}
		case "6":
			value = nil
		default:
			fmt.Printf("%sError:%s Please enter a number between 1 and 7\n", colorRed, colorReset)
			i-- // Don't increment the counter for invalid input
			continue
		}

		result = append(result, value)
	}

	return result
}

// promptForInput shows a prompt and gets user input
func promptForInput(prompt string, required bool) string {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print(prompt + ": ")
		scanner.Scan()
		input := strings.TrimSpace(scanner.Text())

		if input == "" && required {
			fmt.Printf("%sError:%s This field is required\n", colorRed, colorReset)
			continue
		}

		return input
	}
}

// confirmAction asks the user to confirm an action
func confirmAction(prompt string) bool {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Printf("%s? (y/n): ", prompt)
	scanner.Scan()
	input := strings.ToLower(strings.TrimSpace(scanner.Text()))

	return input == "y" || input == "yes"
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
	inputFlag := c.String("input")
	var input string

	// Check if input should be read from stdin
	if inputFlag == "-" {
		fmt.Println("Reading signal input from stdin...")
		scanner := bufio.NewScanner(os.Stdin)
		var inputBuilder strings.Builder
		for scanner.Scan() {
			inputBuilder.WriteString(scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("error reading from stdin: %w", err)
		}
		input = inputBuilder.String()

		// If input is empty, provide a warning
		if strings.TrimSpace(input) == "" {
			fmt.Println("Warning: Empty input received from stdin")
			input = "{}" // Fallback to empty JSON object
		}
	} else {
		// Use the input provided in the flag
		input = inputFlag
	}

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
	argsFlag := c.String("args")
	var args string

	// Check if args should be read from stdin
	if argsFlag == "-" {
		fmt.Println("Reading query args from stdin...")
		scanner := bufio.NewScanner(os.Stdin)
		var argsBuilder strings.Builder
		for scanner.Scan() {
			argsBuilder.WriteString(scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("error reading from stdin: %w", err)
		}
		args = argsBuilder.String()

		// If args is empty, provide a warning
		if strings.TrimSpace(args) == "" {
			fmt.Println("Warning: Empty args received from stdin")
			args = "{}" // Fallback to empty JSON object
		}
	} else {
		// Use the args provided in the flag
		args = argsFlag
	}

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

// inferWorkflowParams tries to infer the parameter structure for a workflow type
// by examining past executions of that type
func inferWorkflowParams(c *cli.Context, config TemporalConfig) error {
	temporalClient, err := getTemporalClient(config)
	if err != nil {
		return fmt.Errorf("failed to create Temporal client: %w", err)
	}
	defer temporalClient.Close()

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	workflowType := c.String("workflow-type")
	limit := c.Int("limit")
	outputAsJSONSchema := c.Bool("json-schema")
	rawOutput := c.Bool("raw")

	if !rawOutput {
		fmt.Printf("Inferring parameter structure for workflow type: %s%s%s\n",
			colorBold, workflowType, colorReset)
	}

	// Find recent workflows of this type
	query := fmt.Sprintf("WorkflowType='%s'", workflowType)
	listRequest := &workflowservice.ListWorkflowExecutionsRequest{
		Query: query,
	}

	resp, err := temporalClient.ListWorkflow(ctx, listRequest)
	if err != nil {
		return fmt.Errorf("failed to list workflows: %w", err)
	}

	if len(resp.Executions) == 0 {
		return fmt.Errorf("no workflows of type '%s' found", workflowType)
	}

	if !rawOutput {
		fmt.Printf("Found %d workflow executions\n", len(resp.Executions))
		fmt.Println("Analyzing recent executions to infer parameter structure...\n")
	}

	// Limit the number of workflows we examine
	numToExamine := limit
	if len(resp.Executions) < numToExamine {
		numToExamine = len(resp.Executions)
	}

	// Map to track distinct parameter structures
	paramStructures := make(map[string]interface{})
	paramExamples := make(map[string]interface{})

	// Examine workflows to extract parameter structures
	for _, execution := range resp.Executions[:numToExamine] {
		workflowID := execution.Execution.WorkflowId
		runID := execution.Execution.RunId

		if !rawOutput {
			fmt.Printf("Examining workflow ID: %s (Run ID: %s)\n", workflowID, runID)
		}

		// Get workflow history to find start event
		iter := temporalClient.GetWorkflowHistory(ctx, workflowID, runID, false, 0)

		// Look for first event (workflow started)
		for iter.HasNext() {
			event, err := iter.Next()
			if err != nil {
				if !rawOutput {
					fmt.Printf("  %sWarning:%s Could not fetch history: %v\n", colorYellow, colorReset, err)
				}
				break
			}

			if event.GetEventType() == enums.EVENT_TYPE_WORKFLOW_EXECUTION_STARTED {
				startedAttrs := event.GetWorkflowExecutionStartedEventAttributes()
				if startedAttrs != nil && startedAttrs.Input != nil {
					payloads := startedAttrs.Input.GetPayloads()
					if len(payloads) > 0 {
						for j, payload := range payloads {
							data := payload.GetData()
							if len(data) > 0 {
								// Try to parse as JSON
								var jsonObj interface{}
								if err := json.Unmarshal(data, &jsonObj); err == nil {
									// Get the structure (just keys for maps, types for other values)
									structureJSON := getStructureJSON(jsonObj)
									structureStr := fmt.Sprintf("%v", structureJSON)

									// Store unique structures
									if _, exists := paramStructures[structureStr]; !exists {
										paramStructures[structureStr] = structureJSON
										paramExamples[structureStr] = jsonObj
									}

									if !rawOutput {
										fmt.Printf("  Found parameter structure (payload %d)\n", j+1)
									}
								} else if !rawOutput {
									fmt.Printf("  %sWarning:%s Parameter is not valid JSON: %v\n",
										colorYellow, colorReset, string(data))
								}
							}
						}
					} else if !rawOutput {
						fmt.Printf("  No input parameters found\n")
					}
				}
				break // Only need to check the first event
			}
		}
	}

	// Display the results based on the requested format
	if len(paramStructures) == 0 {
		if !rawOutput {
			fmt.Printf("\n%s%s==== No Parameter Structures Found ====%s\n",
				colorBold, colorRed, colorReset)
			fmt.Println("Could not determine parameter structure from the examined workflows.")
			fmt.Println("Possible reasons:")
			fmt.Println("  - Workflows don't take parameters")
			fmt.Println("  - Parameters are not in JSON format")
			fmt.Println("  - No workflow history is available")
		}
		return nil
	}

	if outputAsJSONSchema {
		// Generate JSONSchema from the first example we found
		// (or combine multiple schemas if we detected multiple patterns)
		var combinedSchema map[string]interface{}

		if len(paramExamples) == 1 {
			// If we only have one structure, use it directly
			for _, example := range paramExamples {
				combinedSchema = generateJSONSchema(example, workflowType)
				break
			}
		} else {
			// For multiple structures, create a schema with oneOf
			combinedSchema = map[string]interface{}{
				"$schema":     "http://json-schema.org/draft-07/schema#",
				"title":       fmt.Sprintf("%s Parameters", workflowType),
				"description": fmt.Sprintf("Parameter schema for %s workflow", workflowType),
				"oneOf":       []interface{}{},
			}

			oneOf := combinedSchema["oneOf"].([]interface{})
			for _, example := range paramExamples {
				schema := generateJSONSchema(example, "")
				delete(schema, "$schema") // Remove the $schema field from sub-schemas
				oneOf = append(oneOf, schema)
			}
			combinedSchema["oneOf"] = oneOf
		}

		// Output the JSONSchema
		var output []byte
		var err error

		if rawOutput {
			output, err = json.Marshal(combinedSchema)
		} else {
			output, err = json.MarshalIndent(combinedSchema, "", "  ")
		}

		if err != nil {
			return fmt.Errorf("failed to generate JSONSchema: %w", err)
		}

		fmt.Println(string(output))
	} else if !rawOutput {
		// Display the examples in the original format
		fmt.Printf("\n%s%s==== Inferred Parameter Structures ====%s\n",
			colorBold, colorGreen, colorReset)

		fmt.Printf("Found %d distinct parameter structures:\n\n", len(paramExamples))

		i := 1
		for _, example := range paramExamples {
			fmt.Printf("%sStructure %d:%s\n", colorBold, i, colorReset)
			jsonBytes, _ := json.MarshalIndent(example, "", "  ")
			fmt.Println(string(jsonBytes))
			fmt.Println()
			i++
		}

		fmt.Println("You can use these structures as templates when starting new workflows.")
		fmt.Println("To get a JSONSchema, run with the --json-schema flag.")
	}

	return nil
}

// generateJSONSchema converts an example object to a JSONSchema representation
func generateJSONSchema(example interface{}, title string) map[string]interface{} {
	schema := map[string]interface{}{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type":    getJSONType(example),
	}

	if title != "" {
		schema["title"] = fmt.Sprintf("%s Parameters", title)
		schema["description"] = fmt.Sprintf("Parameter schema for %s workflow", title)
	}

	switch v := example.(type) {
	case map[string]interface{}:
		properties := make(map[string]interface{})
		required := []string{}

		for key, val := range v {
			properties[key] = generateJSONSchema(val, "")

			// Assume all fields are required unless they're null
			if val != nil {
				required = append(required, key)
			}
		}

		schema["properties"] = properties
		if len(required) > 0 {
			schema["required"] = required
		}

	case []interface{}:
		if len(v) > 0 {
			// Use the first item to determine the items schema
			schema["items"] = generateJSONSchema(v[0], "")
		} else {
			schema["items"] = map[string]interface{}{"type": "string"}
		}
	}

	return schema
}

// getJSONType returns the JSON schema type for a given value
func getJSONType(v interface{}) string {
	switch v.(type) {
	case map[string]interface{}:
		return "object"
	case []interface{}:
		return "array"
	case string:
		return "string"
	case float64:
		return "number"
	case int:
		return "integer"
	case bool:
		return "boolean"
	case nil:
		return "null"
	default:
		return "string" // fallback
	}
}

// getStructureJSON extracts the structure of a JSON object
// For objects: keeps the keys but replaces values with their types
// For arrays: keeps structure but replaces values with types
// For primitive values: returns the type name
func getStructureJSON(data interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, val := range v {
			result[key] = getStructureJSON(val)
		}
		return result

	case []interface{}:
		if len(v) > 0 {
			// Just use the first element to represent array structure
			return []interface{}{getStructureJSON(v[0])}
		}
		return []interface{}{"empty_array"}

	case string:
		return "string"
	case float64:
		return "number"
	case bool:
		return "boolean"
	case nil:
		return "null"
	default:
		return fmt.Sprintf("unknown_type: %T", v)
	}
}
