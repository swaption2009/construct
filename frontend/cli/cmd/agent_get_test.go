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

func TestAgentGet(t *testing.T) {
	setup := &TestSetup{
		CmpOptions: []cmp.Option{
			cmpopts.IgnoreFields(AgentDisplay{}, "CreatedAt"),
		},
	}

	agentID1 := uuid.New().String()
	agentID2 := uuid.New().String()
	modelID := uuid.New().String()

	setup.RunTests(t, []TestScenario{
		{
			Name:    "success - get agent by name",
			Command: []string{"agent", "get", "coder"},
			SetupMocks: func(mockClient *api_client.MockClient) {
				setupAgentNameLookup(mockClient, "coder", agentID1, modelID)
				setupModelNameLookup(mockClient, "gpt-4", modelID)
				setupAgentGetMock(mockClient, agentID1, "coder", "A helpful coding assistant", "Description for coder", modelID, []string{})
			},
			Expected: TestExpectation{
				DisplayedObjects: &AgentDisplay{
					ID:           agentID1,
					Name:         "coder",
					Description:  "Description for coder",
					Instructions: "A helpful coding assistant",
					Model:        "gpt-4",
				},
			},
		},
		{
			Name:    "success - get agent by ID",
			Command: []string{"agent", "get", agentID1},
			SetupMocks: func(mockClient *api_client.MockClient) {
				setupModelNameLookup(mockClient, "gpt-4", modelID)
				setupAgentGetMock(mockClient, agentID1, "coder", "A helpful coding assistant", "Description for coder", modelID, []string{})
			},
			Expected: TestExpectation{
				DisplayedObjects: &AgentDisplay{
					ID:           agentID1,
					Name:         "coder",
					Description:  "Description for coder",
					Instructions: "A helpful coding assistant",
					Model:        "gpt-4",
				},
			},
		},
		{
			Name:    "success - get agent with JSON output format",
			Command: []string{"agent", "get", "coder", "--output", "json"},
			SetupMocks: func(mockClient *api_client.MockClient) {
				setupAgentNameLookup(mockClient, "coder", agentID1, modelID)
				setupModelNameLookup(mockClient, "gpt-4", modelID)
				setupAgentGetMock(mockClient, agentID1, "coder", "A helpful coding assistant", "Description for coder", modelID, []string{})
			},
			Expected: TestExpectation{
				DisplayFormat: &RenderOptions{
					Format: OutputFormatJSON,
				},
				DisplayedObjects: &AgentDisplay{
					ID:           agentID1,
					Name:         "coder",
					Description:  "Description for coder",
					Instructions: "A helpful coding assistant",
					Model:        "gpt-4",
				},
			},
		},
		{
			Name:    "success - get agent with YAML output format",
			Command: []string{"agent", "get", "coder", "--output", "yaml"},
			SetupMocks: func(mockClient *api_client.MockClient) {
				setupAgentNameLookup(mockClient, "coder", agentID1, modelID)
				setupModelNameLookup(mockClient, "gpt-4", modelID)
				setupAgentGetMock(mockClient, agentID1, "coder", "A helpful coding assistant", "Description for coder", modelID, []string{})
			},
			Expected: TestExpectation{
				DisplayFormat: &RenderOptions{
					Format: OutputFormatYAML,
				},
				DisplayedObjects: &AgentDisplay{
					ID:           agentID1,
					Name:         "coder",
					Description:  "Description for coder",
					Instructions: "A helpful coding assistant",
					Model:        "gpt-4",
				},
			},
		},
		{
			Name:    "error - agent not found by name",
			Command: []string{"agent", "get", "nonexistent"},
			SetupMocks: func(mockClient *api_client.MockClient) {
				mockClient.Agent.EXPECT().ListAgents(
					gomock.Any(),
					&connect.Request[v1.ListAgentsRequest]{
						Msg: &v1.ListAgentsRequest{
							Filter: &v1.ListAgentsRequest_Filter{
								Names: []string{"nonexistent"},
							},
						},
					},
				).Return(&connect.Response[v1.ListAgentsResponse]{
					Msg: &v1.ListAgentsResponse{
						Agents: []*v1.Agent{},
					},
				}, nil)
			},
			Expected: TestExpectation{
				Error: "failed to resolve agent nonexistent: agent nonexistent not found",
			},
		},
		{
			Name:    "error - multiple agents found for name",
			Command: []string{"agent", "get", "duplicate"},
			SetupMocks: func(mockClient *api_client.MockClient) {
				mockClient.Agent.EXPECT().ListAgents(
					gomock.Any(),
					&connect.Request[v1.ListAgentsRequest]{
						Msg: &v1.ListAgentsRequest{
							Filter: &v1.ListAgentsRequest_Filter{
								Names: []string{"duplicate"},
							},
						},
					},
				).Return(&connect.Response[v1.ListAgentsResponse]{
					Msg: &v1.ListAgentsResponse{
						Agents: []*v1.Agent{
							{
								Metadata: &v1.AgentMetadata{
									Id: agentID1,
								},
								Spec: &v1.AgentSpec{
									Name: "duplicate",
								},
							},
							{
								Metadata: &v1.AgentMetadata{
									Id: agentID2,
								},
								Spec: &v1.AgentSpec{
									Name: "duplicate",
								},
							},
						},
					},
				}, nil)
			},
			Expected: TestExpectation{
				Error: "failed to resolve agent duplicate: multiple agents found for duplicate",
			},
		},
		{
			Name:    "error - get agent API failure",
			Command: []string{"agent", "get", agentID1},
			SetupMocks: func(mockClient *api_client.MockClient) {
				mockClient.Agent.EXPECT().GetAgent(
					gomock.Any(),
					&connect.Request[v1.GetAgentRequest]{
						Msg: &v1.GetAgentRequest{Id: agentID1},
					},
				).Return(nil, connect.NewError(connect.CodeInternal, nil))
			},
			Expected: TestExpectation{
				Error: "failed to get agent " + agentID1 + ": internal",
			},
		},
		{
			Name:    "error - list agents API failure during name lookup",
			Command: []string{"agent", "get", "coder"},
			SetupMocks: func(mockClient *api_client.MockClient) {
				mockClient.Agent.EXPECT().ListAgents(
					gomock.Any(),
					&connect.Request[v1.ListAgentsRequest]{
						Msg: &v1.ListAgentsRequest{
							Filter: &v1.ListAgentsRequest_Filter{
								Names: []string{"coder"},
							},
						},
					},
				).Return(nil, connect.NewError(connect.CodeInternal, nil))
			},
			Expected: TestExpectation{
				Error: "failed to resolve agent coder: failed to list agents: internal",
			},
		},
	})
}

func setupAgentNameLookup(mockClient *api_client.MockClient, agentName, agentID string, modelID string) {
	mockClient.Agent.EXPECT().ListAgents(
		gomock.Any(),
		&connect.Request[v1.ListAgentsRequest]{
			Msg: &v1.ListAgentsRequest{
				Filter: &v1.ListAgentsRequest_Filter{
					Names: []string{agentName},
				},
			},
		},
	).Return(&connect.Response[v1.ListAgentsResponse]{
		Msg: &v1.ListAgentsResponse{
			Agents: []*v1.Agent{
				{
					Metadata: &v1.AgentMetadata{
						Id: agentID,
					},
					Spec: &v1.AgentSpec{
						Name:    agentName,
						ModelId: modelID,
					},
				},
			},
		},
	}, nil)
}

func setupAgentGetMock(mockClient *api_client.MockClient, agentID, name, instructions, description, modelID string, delegateIDs []string) {
	agent := &v1.Agent{
		Metadata: &v1.AgentMetadata{
			Id: agentID,
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

	mockClient.Agent.EXPECT().GetAgent(
		gomock.Any(),
		&connect.Request[v1.GetAgentRequest]{
			Msg: &v1.GetAgentRequest{Id: agentID},
		},
	).Return(&connect.Response[v1.GetAgentResponse]{
		Msg: &v1.GetAgentResponse{
			Agent: agent,
		},
	}, nil)
}
