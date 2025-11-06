package cmd

import (
	"fmt"

	"connectrpc.com/connect"
	v1 "github.com/furisto/construct/api/go/v1"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

type taskListOptions struct {
	Agent         string
	Limit         int32
	RenderOptions RenderOptions
}

func NewTaskListCmd() *cobra.Command {
	var options taskListOptions

	cmd := &cobra.Command{
		Use:     "list [flags]",
		Short:   "List all tasks",
		Aliases: []string{"ls"},
		Example: `  # List all recent tasks
  construct task list

  # List tasks assigned to the 'coder' agent, in JSON format
  construct task ls --agent "coder" --output json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := getAPIClient(cmd.Context())

			filter := &v1.ListTasksRequest_Filter{}

			if options.Agent != "" {
				agentID := options.Agent
				_, err := uuid.Parse(agentID)
				if err != nil {
					resolvedID, err := getAgentID(cmd.Context(), client, agentID)
					if err != nil {
						return fmt.Errorf("failed to resolve agent %s: %w", agentID, err)
					}
					agentID = resolvedID
				}
				filter.AgentId = &agentID
			}

			req := &connect.Request[v1.ListTasksRequest]{
				Msg: &v1.ListTasksRequest{
					Filter:   filter,
					PageSize: &options.Limit,
				},
			}

			resp, err := client.Task().ListTasks(cmd.Context(), req)
			if err != nil {
				return fmt.Errorf("failed to list tasks: %w", err)
			}

			displayTasks := make([]*DisplayTask, len(resp.Msg.Tasks))
			for i, task := range resp.Msg.Tasks {
				displayTasks[i] = ConvertTaskToDisplay(task)
			}

			return getRenderer(cmd.Context()).Render(displayTasks, &options.RenderOptions)
		},
	}

	cmd.Flags().StringVarP(&options.Agent, "agent", "a", "", "Filter tasks by the agent assigned to them")
	cmd.Flags().Int32VarP(&options.Limit, "limit", "l", 0, "Limit the number of results returned")
	addRenderOptions(cmd, &options.RenderOptions)
	return cmd
}
