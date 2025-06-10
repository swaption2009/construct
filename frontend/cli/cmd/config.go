package cmd

import (
	"github.com/spf13/cobra"
)

var configOptions = []string{
	"task.default-agent",
	"agent.default-model",
}

func NewConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "config",
		Short:   "Manage configuration",
		GroupID: "system",
	}

	cmd.AddCommand(NewConfigSetCmd())
	cmd.AddCommand(NewConfigGetCmd())
	cmd.AddCommand(NewConfigUnsetCmd())
	cmd.AddCommand(NewConfigDescribeCmd())
	cmd.AddCommand(NewConfigListCmd())

	return cmd
}
