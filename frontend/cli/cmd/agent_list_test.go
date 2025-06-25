package cmd

import (
	"testing"

	"connectrpc.com/connect"
	api_client "github.com/furisto/construct/api/go/client"
	v1 "github.com/furisto/construct/api/go/v1"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

func TestAgentList(t *testing.T) {
	setup := &TestSetup{
		CmpOptions: []cmp.Option{
			cmpopts.IgnoreFields(AgentDisplay{}, "CreatedAt"),
		},
	}

	agentID1 := uuid.New().String()
	agentID2 := uuid.New().String()
	modelID1 := uuid.New().String()
	modelID2 := uuid.New().String()

	setup.RunTests(t, []TestScenario{
		{
			Name:    "success - list all agents",
			Command: []string{"agent", "list"},
			SetupMocks: func(mockClient *api_client.MockClient) {
				setupAgentListRequestMock(mockClient, nil, []*v1.Agent{
					createTestAgent(agentID1, "coder", "A helpful coding assistant", "Description for coder", modelID1),
					createTestAgent(agentID2, "reviewer", "A code reviewer", "Description for reviewer", modelID2),
				})
				setupModelNameLookup(mockClient, "gpt-4", modelID1)
				setupModelNameLookup(mockClient, "claude-4", modelID2)
			},
			Expected: TestExpectation{
				DisplayedObjects: []*AgentDisplay{
					{
						ID:           agentID1,
						Name:         "coder",
						Description:  "Description for coder",
						Instructions: "A helpful coding assistant",
						Model:        "gpt-4",
					},
					{
						ID:           agentID2,
						Name:         "reviewer",
						Description:  "Description for reviewer",
						Instructions: "A code reviewer",
						Model:        "claude-4",
					},
				},
			},
		},
		{
			Name:    "success - list agents with JSON output",
			Command: []string{"agent", "list", "--output", "json"},
			SetupMocks: func(mockClient *api_client.MockClient) {
				setupAgentListRequestMock(mockClient, nil, []*v1.Agent{
					createTestAgent(agentID1, "coder", "A helpful coding assistant", "Description for coder", modelID1),
				})
				setupModelNameLookup(mockClient, "gpt-4", modelID1)
			},
			Expected: TestExpectation{
				DisplayFormat: &RenderOptions{
					Format: OutputFormatJSON,
				},
				DisplayedObjects: []*AgentDisplay{
					{
						ID:           agentID1,
						Name:         "coder",
						Description:  "Description for coder",
						Instructions: "A helpful coding assistant",
						Model:        "gpt-4",
					},
				},
			},
		},
		{
			Name:    "success - filter agents by model ID (client-side filtering)",
			Command: []string{"agent", "list", "--model", modelID1},
			SetupMocks: func(mockClient *api_client.MockClient) {
				filter := &v1.ListAgentsRequest_Filter{}
				setupAgentListRequestMock(mockClient, filter, []*v1.Agent{
					createTestAgent(agentID1, "coder", "A helpful coding assistant", "Description for coder", modelID1),
				})
				setupModelNameLookup(mockClient, "gpt-4", modelID1)
			},
			Expected: TestExpectation{
				DisplayedObjects: []*AgentDisplay{
					{
						ID:           agentID1,
						Name:         "coder",
						Description:  "Description for coder",
						Instructions: "A helpful coding assistant",
						Model:        "gpt-4",
					},
				},
			},
		},
		{
			Name:    "success - filter agents by model name",
			Command: []string{"agent", "list", "--model", "claude-4"},
			SetupMocks: func(mockClient *api_client.MockClient) {
				setupModelLookupMock(mockClient, "claude-4", modelID1)
				filter := &v1.ListAgentsRequest_Filter{}
				setupAgentListRequestMock(mockClient, filter, []*v1.Agent{
					createTestAgent(agentID1, "coder", "A helpful coding assistant", "Description for coder", modelID1),
				})
				setupModelNameLookup(mockClient, "claude-4", modelID1)
			},
			Expected: TestExpectation{
				DisplayedObjects: []*AgentDisplay{
					{
						ID:           agentID1,
						Name:         "coder",
						Description:  "Description for coder",
						Instructions: "A helpful coding assistant",
						Model:        "claude-4",
					},
				},
			},
		},
		{
			Name:    "success - filter agents by multiple models",
			Command: []string{"agent", "list", "--model", modelID1, "--model", "claude-4"},
			SetupMocks: func(mockClient *api_client.MockClient) {
				setupModelLookupMock(mockClient, "claude-4", modelID2)
				filter := &v1.ListAgentsRequest_Filter{}
				setupAgentListRequestMock(mockClient, filter, []*v1.Agent{
					createTestAgent(agentID1, "coder", "A helpful coding assistant", "Description for coder", modelID1),
					createTestAgent(agentID2, "reviewer", "A code reviewer", "Description for reviewer", modelID2),
				})
				setupModelNameLookup(mockClient, "gpt-4", modelID1)
				setupModelNameLookup(mockClient, "claude-4", modelID2)
			},
			Expected: TestExpectation{
				DisplayedObjects: []*AgentDisplay{
					{
						ID:           agentID1,
						Name:         "coder",
						Description:  "Description for coder",
						Instructions: "A helpful coding assistant",
						Model:        "gpt-4",
					},
					{
						ID:           agentID2,
						Name:         "reviewer",
						Description:  "Description for reviewer",
						Instructions: "A code reviewer",
						Model:        "claude-4",
					},
				},
			},
		},
		{
			Name:    "success - filter agents by name",
			Command: []string{"agent", "list", "--name", "coder"},
			SetupMocks: func(mockClient *api_client.MockClient) {
				filter := &v1.ListAgentsRequest_Filter{
					Names: []string{"coder"},
				}
				setupAgentListRequestMock(mockClient, filter, []*v1.Agent{
					createTestAgent(agentID1, "coder", "A helpful coding assistant", "Description for coder", modelID1),
				})
				setupModelNameLookup(mockClient, "gpt-4", modelID1)
			},
			Expected: TestExpectation{
				DisplayedObjects: []*AgentDisplay{
					{
						ID:           agentID1,
						Name:         "coder",
						Description:  "Description for coder",
						Instructions: "A helpful coding assistant",
						Model:        "gpt-4",
					},
				},
			},
		},
		{
			Name:    "success - filter agents by multiple names",
			Command: []string{"agent", "list", "--name", "coder", "--name", "reviewer"},
			SetupMocks: func(mockClient *api_client.MockClient) {
				filter := &v1.ListAgentsRequest_Filter{
					Names: []string{"coder", "reviewer"},
				}
				setupAgentListRequestMock(mockClient, filter, []*v1.Agent{
					createTestAgent(agentID1, "coder", "A helpful coding assistant", "Description for coder", modelID1),
					createTestAgent(agentID2, "reviewer", "A code reviewer", "Description for reviewer", modelID2),
				})
				setupModelNameLookup(mockClient, "gpt-4", modelID1)
				setupModelNameLookup(mockClient, "claude-4", modelID2)
			},
			Expected: TestExpectation{
				DisplayedObjects: []*AgentDisplay{
					{
						ID:           agentID1,
						Name:         "coder",
						Description:  "Description for coder",
						Instructions: "A helpful coding assistant",
						Model:        "gpt-4",
					},
					{
						ID:           agentID2,
						Name:         "reviewer",
						Description:  "Description for reviewer",
						Instructions: "A code reviewer",
						Model:        "claude-4",
					},
				},
			},
		},
		{
			Name:    "success - combined filters (model and name)",
			Command: []string{"agent", "list", "--model", modelID1, "--name", "coder"},
			SetupMocks: func(mockClient *api_client.MockClient) {
				filter := &v1.ListAgentsRequest_Filter{
					Names: []string{"coder"},
				}
				setupAgentListRequestMock(mockClient, filter, []*v1.Agent{
					createTestAgent(agentID1, "coder", "A helpful coding assistant", "Description for coder", modelID1),
				})
				setupModelNameLookup(mockClient, "gpt-4", modelID1)
			},
			Expected: TestExpectation{
				DisplayedObjects: []*AgentDisplay{
					{
						ID:           agentID1,
						Name:         "coder",
						Description:  "Description for coder",
						Instructions: "A helpful coding assistant",
						Model:        "gpt-4",
					},
				},
			},
		},
		{
			Name:    "success - empty result set",
			Command: []string{"agent", "list"},
			SetupMocks: func(mockClient *api_client.MockClient) {
				setupAgentListRequestMock(mockClient, nil, []*v1.Agent{})
			},
			Expected: TestExpectation{
				DisplayedObjects: []*AgentDisplay{},
			},
		},
		{
			Name:    "error - list agents API failure",
			Command: []string{"agent", "list"},
			SetupMocks: func(mockClient *api_client.MockClient) {
				mockClient.Agent.EXPECT().ListAgents(
					gomock.Any(),
					&connect.Request[v1.ListAgentsRequest]{
						Msg: &v1.ListAgentsRequest{
							Filter: &v1.ListAgentsRequest_Filter{},
						},
					},
				).Return(nil, connect.NewError(connect.CodeInternal, nil))
			},
			Expected: TestExpectation{
				Error: "failed to list agents: internal",
			},
		},
		{
			Name:    "error - model not found",
			Command: []string{"agent", "list", "--model", "nonexistent-model"},
			SetupMocks: func(mockClient *api_client.MockClient) {
				mockClient.Model.EXPECT().ListModels(
					gomock.Any(),
					&connect.Request[v1.ListModelsRequest]{
						Msg: &v1.ListModelsRequest{
							Filter: &v1.ListModelsRequest_Filter{
								Names: []string{"nonexistent-model"},
							},
						},
					},
				).Return(&connect.Response[v1.ListModelsResponse]{
					Msg: &v1.ListModelsResponse{
						Models: []*v1.Model{},
					},
				}, nil)
			},
			Expected: TestExpectation{
				Error: "failed to resolve model nonexistent-model: model nonexistent-model not found",
			},
		},
		{
			Name:    "error - multiple models found for name",
			Command: []string{"agent", "list", "--model", "duplicate-model"},
			SetupMocks: func(mockClient *api_client.MockClient) {
				mockClient.Model.EXPECT().ListModels(
					gomock.Any(),
					&connect.Request[v1.ListModelsRequest]{
						Msg: &v1.ListModelsRequest{
							Filter: &v1.ListModelsRequest_Filter{
								Names: []string{"duplicate-model"},
							},
						},
					},
				).Return(&connect.Response[v1.ListModelsResponse]{
					Msg: &v1.ListModelsResponse{
						Models: []*v1.Model{
							{
								Metadata: &v1.ModelMetadata{Id: modelID1},
								Spec:     &v1.ModelSpec{Name: "duplicate-model"},
							},
							{
								Metadata: &v1.ModelMetadata{Id: modelID2},
								Spec:     &v1.ModelSpec{Name: "duplicate-model"},
							},
						},
					},
				}, nil)
			},
			Expected: TestExpectation{
				Error: "failed to resolve model duplicate-model: multiple models found for duplicate-model",
			},
		},
		{
			Name:    "error - list models API failure during name lookup",
			Command: []string{"agent", "list", "--model", "claude-4"},
			SetupMocks: func(mockClient *api_client.MockClient) {
				mockClient.Model.EXPECT().ListModels(
					gomock.Any(),
					&connect.Request[v1.ListModelsRequest]{
						Msg: &v1.ListModelsRequest{
							Filter: &v1.ListModelsRequest_Filter{
								Names: []string{"claude-4"},
							},
						},
					},
				).Return(nil, connect.NewError(connect.CodeInternal, nil))
			},
			Expected: TestExpectation{
				Error: "failed to resolve model claude-4: failed to list models: internal",
			},
		},
	})
}

func setupAgentListRequestMock(mockClient *api_client.MockClient, filter *v1.ListAgentsRequest_Filter, agents []*v1.Agent) {
	if filter == nil {
		filter = &v1.ListAgentsRequest_Filter{}
	}

	mockClient.Agent.EXPECT().ListAgents(
		gomock.Any(),
		&connect.Request[v1.ListAgentsRequest]{
			Msg: &v1.ListAgentsRequest{
				Filter: filter,
			},
		},
	).Return(&connect.Response[v1.ListAgentsResponse]{
		Msg: &v1.ListAgentsResponse{
			Agents: agents,
		},
	}, nil)
}

func createTestAgent(id, name, instructions, description, modelID string) *v1.Agent {
	agent := &v1.Agent{
		Metadata: &v1.AgentMetadata{
			Id: id,
		},
		Spec: &v1.AgentSpec{
			Name:         name,
			Instructions: instructions,
			ModelId:      modelID,
		},
	}

	if description != "" {
		agent.Spec.Description = description
	}

	return agent
}
