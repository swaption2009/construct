package agent

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/furisto/construct/backend/memory"
	memory_message "github.com/furisto/construct/backend/memory/message"
	"github.com/furisto/construct/backend/memory/schema/types"
	memory_task "github.com/furisto/construct/backend/memory/task"
	"github.com/furisto/construct/backend/model"
	"github.com/google/uuid"
)

const (
	MaxTitleLength = 80 // characters
	MaxRetries     = 2  // attempts to get shorter title
)

type TitleGenerator struct {
	memory          *memory.Client
	providerFactory *ModelProviderFactory
}

func NewTitleGenerator(memory *memory.Client, providerFactory *ModelProviderFactory) *TitleGenerator {
	return &TitleGenerator{
		memory:          memory,
		providerFactory: providerFactory,
	}
}

func (g *TitleGenerator) GenerateTitle(ctx context.Context, taskID uuid.UUID) error {
	_, agent, err := g.fetchTaskWithAgent(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to fetch task: %w", err)
	}

	messages, err := g.memory.Message.Query().
		Where(memory_message.TaskIDEQ(taskID)).
		Order(memory_message.ByCreateTime()).
		Limit(5).
		All(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch messages: %w", err)
	}

	if !hasUserMessage(messages) {
		slog.DebugContext(ctx, "no user messages, skipping title generation", "task_id", taskID)
		return nil
	}

	modelProvider, err := g.providerFactory.CreateClient(ctx, agent.Edges.Model.ModelProviderID)
	if err != nil {
		return fmt.Errorf("failed to create model provider: %w", err)
	}

	modelMessages, err := g.buildMessageHistory(messages)
	if err != nil {
		return fmt.Errorf("failed to build message history: %w", err)
	}

	systemPrompt := `You are a title generator for development tasks and conversations. Generate a concise, descriptive title based on the user's request and conversation context.

CRITICAL RULES:
- Maximum 8 words
- Maximum 80 characters
- Start immediately with the title (no preamble)
- Use action verbs when describing tasks
- Be specific about technology/domain when relevant
- NO quotes, markdown, or extra punctuation
- NO meta-commentary or explanations

GOOD EXAMPLES:
- "Implement JWT authentication for API"
- "Fix memory leak in worker pool"
- "Add TypeScript support to build"
- "Debug race condition in cache"
- "Refactor database connection pooling"

BAD EXAMPLES (too vague):
- "Help with code"
- "Fix issue"
- "Question about app"

BAD EXAMPLES (too long):
- "Implement a comprehensive user authentication system with JWT tokens and refresh functionality"

For the given conversation, generate ONLY the title starting with a quote:`

	title, err := g.generateTitleWithRetry(ctx, modelProvider, systemPrompt, modelMessages, 0)
	if err != nil {
		return fmt.Errorf("failed to generate title: %w", err)
	}

	_, err = memory.Transaction(ctx, g.memory, func(tx *memory.Client) (*memory.Task, error) {
		return tx.Task.UpdateOneID(taskID).SetDescription(title).Save(ctx)
	})
	if err != nil {
		return fmt.Errorf("failed to save title: %w", err)
	}

	slog.InfoContext(ctx, "generated title for task", "task_id", taskID, "title", title)
	return nil
}

func (g *TitleGenerator) generateTitleWithRetry(
	ctx context.Context,
	provider model.ModelProvider,
	systemPrompt string,
	messages []*model.Message,
	attempt int,
) (string, error) {
	messagesWithPrefill := append(messages, &model.Message{
		Source: model.MessageSourceModel,
		Content: []model.ContentBlock{
			&model.TextBlock{Text: "\""},
		},
	})

	var anthropicProvider *model.AnthropicProvider
	anthropicProvider, ok := provider.(*model.AnthropicProvider)
	if !ok {
		return "", fmt.Errorf("provider is not an Anthropic provider")
	}

	budgetModel := anthropicProvider.BudgetModel()

	response, err := provider.InvokeModel(
		ctx,
		budgetModel,
		systemPrompt,
		messagesWithPrefill,
	)
	if err != nil {
		return "", err
	}

	title := extractTitle(response.Content)
	if title == "" {
		return "", fmt.Errorf("model returned empty title")
	}

	if len(title) > MaxTitleLength && attempt < MaxRetries {
		slog.DebugContext(ctx, "title too long, retrying", "attempt", attempt+1, "length", len(title), "title", title)

		messages = append(messages,
			&model.Message{
				Source: model.MessageSourceModel,
				Content: []model.ContentBlock{
					&model.TextBlock{Text: title},
				},
			},
			&model.Message{
				Source: model.MessageSourceUser,
				Content: []model.ContentBlock{
					&model.TextBlock{
						Text: "That title is too long. Please provide a shorter version (max 8 words, ~80 characters).",
					},
				},
			},
		)

		return g.generateTitleWithRetry(ctx, provider, systemPrompt, messages, attempt+1)
	}

	if len(title) > MaxTitleLength {
		slog.WarnContext(ctx, "title still too long after retries, truncating", "original_length", len(title))
		title = title[:MaxTitleLength-3] + "..."
	}

	return title, nil
}

func (g *TitleGenerator) fetchTaskWithAgent(ctx context.Context, taskID uuid.UUID) (*memory.Task, *memory.Agent, error) {
	task, err := g.memory.Task.Query().
		Where(memory_task.IDEQ(taskID)).
		WithAgent(func(query *memory.AgentQuery) {
			query.WithModel()
		}).
		Only(ctx)

	if err != nil {
		return nil, nil, err
	}

	if task.Edges.Agent == nil {
		return nil, nil, fmt.Errorf("no agent associated with task: %s", taskID)
	}

	return task, task.Edges.Agent, nil
}

func (g *TitleGenerator) buildMessageHistory(messages []*memory.Message) ([]*model.Message, error) {
	modelMessages := make([]*model.Message, 0, len(messages))

	for _, msg := range messages {
		if msg.Source == types.MessageSourceUser || msg.Source == types.MessageSourceAssistant {
			modelMsg, err := ConvertMemoryMessageToModel(msg)
			if err != nil {
				return nil, err
			}
			modelMessages = append(modelMessages, modelMsg)
		}
	}

	return modelMessages, nil
}

func extractTitle(content []model.ContentBlock) string {
	var title string
	for _, block := range content {
		if textBlock, ok := block.(*model.TextBlock); ok {
			title = textBlock.Text
			break
		}
	}

	title = strings.Trim(title, "\" \n\t\r'`")
	title = strings.TrimSpace(title)

	return title
}

func hasUserMessage(messages []*memory.Message) bool {
	for _, msg := range messages {
		if msg.Source == types.MessageSourceUser {
			return true
		}
	}
	return false
}
