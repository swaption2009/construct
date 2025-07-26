package communication

import (
	"context"
	"fmt"

	"github.com/furisto/construct/backend/memory"
	"github.com/furisto/construct/backend/memory/agent"
	"github.com/furisto/construct/backend/memory/schema/types"
	"github.com/furisto/construct/backend/tool/base"
	"github.com/google/uuid"
)

type HandoffInput struct {
	TaskID          uuid.UUID
	CurrentAgentID  uuid.UUID
	RequestedAgent  string
	HandoverMessage string
}

func Handoff(ctx context.Context, db *memory.Client, input *HandoffInput) error {
	if input.TaskID == uuid.Nil {
		return base.NewCustomError("task_id is required", []string{
			"Ensure the task ID is properly set in the session context",
		})
	}
	if input.CurrentAgentID == uuid.Nil {
		return base.NewCustomError("current_agent_id is required", []string{
			"Ensure the current agent ID is properly set in the session context",
		})
	}
	if input.RequestedAgent == "" {
		return base.NewCustomError("requested_agent is required", []string{
			"Provide the name of the agent to handoff to",
		})
	}

	_, err := memory.Transaction(ctx, db, func(tx *memory.Client) (*any, error) {
		currentAgent, err := tx.Agent.Query().Where(agent.IDEQ(input.CurrentAgentID)).WithModel().Only(ctx)
		if err != nil {
			return nil, base.NewCustomError(fmt.Sprintf("failed to get current agent: %v", memory.SanitizeError(err)), []string{
				"This is likely due to a bug in the system or the agent being deleted in the meantime. Retrying this operation probably won't help.",
				"Ask the user how to proceed",
			})
		}

		requestedAgent, err := tx.Agent.Query().Where(agent.NameEQ(input.RequestedAgent)).First(ctx)
		if err != nil {
			if memory.IsNotFound(err) {
				return nil, base.NewCustomError(fmt.Sprintf("agent %s does not exist", input.RequestedAgent), []string{
					"Check the agent name and try again",
				})
			}
			return nil, err
		}

		if requestedAgent.ID == currentAgent.ID {
			return nil, base.NewCustomError("agent cannot handoff to itself", []string{
				fmt.Sprintf("You are the source %s agent and cannot handoff to yourself", currentAgent.Name),
			})
		}

		task, err := tx.Task.Get(ctx, input.TaskID)
		if err != nil {
			return nil, base.NewCustomError(fmt.Sprintf("failed to get task: %v", memory.SanitizeError(err)), []string{
				"This is likely due to a bug in the system or the task being deleted in the meantime. Retrying this operation probably won't help.",
				"Ask the user how to proceed",
			})
		}

		_, err = task.Update().SetAgent(requestedAgent).Save(ctx)
		if err != nil {
			return nil, err
		}

		message := fmt.Sprintf("The %s agent has performed a handoff to you, the %s agent.\n\n", currentAgent.Name, input.RequestedAgent)
		if len(input.HandoverMessage) > 0 {
			message += fmt.Sprintf("It has left the following instructions for you:\n%s", input.HandoverMessage)
		}

		_, err = tx.Message.Create().
			SetContent(&types.MessageContent{
				Blocks: []types.MessageBlock{
					{
						Kind:    types.MessageBlockKindText,
						Payload: message,
					},
				},
			}).
			SetTask(task).
			SetAgent(currentAgent).
			SetModel(currentAgent.Edges.Model).
			SetSource(types.MessageSourceAssistant).
			Save(ctx)
		return nil, err
	})

	return err
}
