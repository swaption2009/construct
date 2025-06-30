package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"connectrpc.com/connect"
	v1 "github.com/furisto/construct/api/go/v1"
	"github.com/furisto/construct/frontend/cli/pkg/fail"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
	"gopkg.in/yaml.v3"
)

type askOutputFormat string

const (
	askOutputFormatText askOutputFormat = "text"
	askOutputFormatJSON askOutputFormat = "json"
	askOutputFormatYAML askOutputFormat = "yaml"
)

func (e *askOutputFormat) String() string {
	if e == nil || *e == "" {
		return string(askOutputFormatText)
	}
	return string(*e)
}

func (e *askOutputFormat) Set(v string) error {
	switch v {
	case "text", "json", "yaml":
		*e = askOutputFormat(v)
		return nil
	default:
		return errors.New(`must be one of "text", "json", or "yaml"`)
	}
}

func (e *askOutputFormat) Type() string {
	return "format"
}

type askOptions struct {
	Agent     string
	Workspace string
	MaxTurns  int
	Continue  bool
	Files     []string
	Format    askOutputFormat
}

func NewAskCmd() *cobra.Command {
	options := askOptions{
		Format: askOutputFormatText,
	}

	cmd := &cobra.Command{
		Use:     "ask [question]",
		Short:   "Ask a question to the AI",
		Args:    cobra.MaximumNArgs(1),
		GroupID: "core",
		Example: `  # Simple question
  construct ask "What is 2+2?"

  # Use a specific agent
  construct ask "Review this code for security issues" --agent security-reviewer

  # Include files as context
  construct ask "What does this code do?" --file main.go --file utils.go

  # Pipe input with question and file context
  cat main.go | construct ask "What does this code do?" --file config.yaml

  # Give agent more turns for complex tasks
  construct ask "Debug why the build is failing" --max-turns 10

  # Get JSON output for scripting
  construct ask "List all Go files" --output json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var question string
			if len(args) > 0 {
				question = args[0]
			}
			return fail.HandleError(ask(cmd.Context(), cmd, options, question))
		},
	}

	cmd.Flags().StringVarP(&options.Agent, "agent", "a", "", "The agent to use (name or ID)")
	cmd.Flags().StringVarP(&options.Workspace, "workspace", "w", "", "The workspace directory")
	cmd.Flags().IntVar(&options.MaxTurns, "max-turns", 5, "Maximum number of turns for the conversation")
	cmd.Flags().StringSliceVarP(&options.Files, "file", "f", []string{}, "Files to include as context (can be used multiple times)")
	cmd.Flags().BoolVarP(&options.Continue, "continue", "c", false, "Continue the previous task")
	cmd.Flags().VarP(&options.Format, "output", "o", "The format to output the result in")

	return cmd
}

func ask(ctx context.Context, cmd *cobra.Command, options askOptions, question string) error {
	client := getAPIClient(ctx)

	question, err := getQuestion(question, cmd.InOrStdin())
	if err != nil {
		return err
	}

	message, err := buildMessage(question, options.Files, getFileSystem(ctx))
	if err != nil {
		return err
	}

	workspace := options.Workspace
	if workspace == "" {
		workspace, err = os.Getwd()
		if err != nil {
			return err
		}
	}

	agentID, err := getAgentID(ctx, client, options.Agent)
	if err != nil {
		return err
	}

	var task *v1.Task
	if options.Continue {
		// tasks, err := client.Task().ListTasks(ctx, &connect.Request[v1.ListTasksRequest]{
		// 	Msg: &v1.ListTasksRequest_Filter{
		// 		AgentId: 	conv.Ptr(agentID),
		// 	},
		// })
	} else {
		taskResp, err := client.Task().CreateTask(ctx, &connect.Request[v1.CreateTaskRequest]{
			Msg: &v1.CreateTaskRequest{
				AgentId:          agentID,
				ProjectDirectory: workspace,
			},
		})
		if err != nil {
			return fmt.Errorf("failed to create task: %w", err)
		}
		task = taskResp.Msg.Task
	}

	_, err = client.Message().CreateMessage(ctx, &connect.Request[v1.CreateMessageRequest]{
		Msg: &v1.CreateMessageRequest{
			TaskId: task.Metadata.Id,
			Content: []*v1.MessagePart{
				{
					Data: &v1.MessagePart_Text_{
						Text: &v1.MessagePart_Text{
							Content: message,
						},
					},
				},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	stream, err := client.Task().Subscribe(ctx, &connect.Request[v1.SubscribeRequest]{
		Msg: &v1.SubscribeRequest{
			TaskId: task.Metadata.Id,
		},
	})

	if err != nil {
		return fmt.Errorf("failed to subscribe to task: %w", err)
	}

	for stream.Receive() {
		message := stream.Msg().Message
		switch options.Format {
		case askOutputFormatText:
			for _, part := range message.Spec.Content {
				switch partData := part.Data.(type) {
				case *v1.MessagePart_Text_:
					fmt.Fprintln(cmd.OutOrStdout(), partData.Text.Content)
				case *v1.MessagePart_ToolResult_:
					fmt.Fprintln(cmd.OutOrStdout(), partData.ToolResult.Result)
				}
			}
		case askOutputFormatJSON:
			jsonBytes, err := toJSON(message)
			if err != nil {
				return fmt.Errorf("failed to marshal to JSON: %w", err)
			}

			var buf bytes.Buffer
			json.Indent(&buf, jsonBytes, "", "  ")
			fmt.Fprintln(cmd.OutOrStdout(), buf.String())
		case askOutputFormatYAML:
			jsonBytes, err := toJSON(message)
			if err != nil {
				return fmt.Errorf("failed to marshal to JSON: %w", err)
			}

			var jsonData interface{}
			if err := json.Unmarshal(jsonBytes, &jsonData); err != nil {
				return fmt.Errorf("failed to unmarshal JSON: %w", err)
			}

			if err := yaml.NewEncoder(cmd.OutOrStdout()).Encode(jsonData); err != nil {
				return fmt.Errorf("failed to marshal to YAML: %w", err)
			}
		default:
			return fmt.Errorf("unsupported format: %s", options.Format)
		}
	}

	if err := stream.Err(); err != nil {
		return fmt.Errorf("stream error: %w", err)
	}

	return nil
}

func getQuestion(question string, stdin io.Reader) (string, error) {
	var stdinContent string

	stat, err := os.Stdin.Stat()
	if err == nil && (stat.Mode()&os.ModeCharDevice) == 0 {
		content, err := io.ReadAll(stdin)
		if err != nil {
			return "", fmt.Errorf("failed to read input from stdin: %w", err)
		}
		stdinContent = strings.TrimSpace(string(content))
	}

	switch {
	case stdinContent != "" && question != "":
		return fmt.Sprintf("%s\n\n%s", question, stdinContent), nil
	case stdinContent != "":
		return stdinContent, nil
	case question != "":
		return question, nil
	}

	return "", fmt.Errorf("no question provided - provide as argument or pipe via stdin")
}

func buildMessage(question string, files []string, fs afero.Fs) (string, error) {
	if len(files) == 0 {
		return question, nil
	}

	var builder strings.Builder
	builder.WriteString(question + "\n")
	builder.WriteString("--- File Context ---\n")

	for i, filepath := range files {
		content, err := fs.Open(filepath)
		if err != nil {
			return "", fmt.Errorf("failed to read file %s: %w", filepath, err)
		}

		builder.WriteString(fmt.Sprintf("### %s\n```\n", filepath))
		_, err = io.Copy(&builder, content)
		content.Close()
		if err != nil {
			return "", fmt.Errorf("failed to read file %s: %w", filepath, err)
		}
		if i < len(files)-1 {
			builder.WriteString("\n```\n\n")
		} else {
			builder.WriteString("\n```")
		}
	}

	return builder.String(), nil
}

func toJSON(message *v1.Message) ([]byte, error) {
	marshaler := protojson.MarshalOptions{
		Multiline:       true,
		Indent:          "  ",
		UseProtoNames:   false,
		EmitUnpopulated: false,
	}

	return marshaler.Marshal(message)
}
