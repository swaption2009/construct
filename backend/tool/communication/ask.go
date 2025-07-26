package communication

import (
	"context"

	"github.com/spf13/afero"
)

// AskUserInput represents the input for asking the user
type AskUserInput struct {
	Question string   `json:"question"`
	Options  []string `json:"options,omitempty"`
}

// AskUserResult represents the result of asking the user
type AskUserResult struct {
	UserResponse   string `json:"user_response"`
	SelectedOption string `json:"selected_option,omitempty"`
}

// AskUserTool implements the core user interaction functionality
type AskUserTool struct{}

// NewAskUserTool creates a new instance of the ask user tool
func NewAskUserTool() *AskUserTool {
	return &AskUserTool{}
}

// Name returns the tool name
func (t *AskUserTool) Name() string {
	return "ask_user"
}

// Description returns the tool description
func (t *AskUserTool) Description() string {
	return "Initiates interactive communication with the user to gather additional information or clarification."
}

// Execute runs the ask user operation
func (t *AskUserTool) Execute(ctx context.Context, fs afero.Fs, input interface{}) (interface{}, error) {
	// TODO: Implement core ask user logic here
	// This will be extracted from the current ask_user.go in Phase 3
	return nil, nil
}
