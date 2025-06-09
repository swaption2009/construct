package cmd

import (
	"testing"
	"time"

	"connectrpc.com/connect"
	api_client "github.com/furisto/construct/api/go/client"
	v1 "github.com/furisto/construct/api/go/v1"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

func TestModelCreate(t *testing.T) {
	setup := &TestSetup{}

	modelID1 := uuid.New().String()
	modelProviderID1 := uuid.New().String()
	createdAt := time.Now()

	setup.RunTests(t, []TestScenario{
		{
			Name:    "success - create model with provider by name",
			Command: []string{"model", "create", "gpt-4", "--model-provider", "openai-dev", "--context-window", "8192"},
			SetupMocks: func(mockClient *api_client.MockClient) {
				setupModelProviderLookupForCreateMock(mockClient, "openai-dev", modelProviderID1)
				setupModelCreateMock(mockClient, "gpt-4", modelProviderID1, 8192, modelID1, createdAt)
			},
			Expected: TestExpectation{
				Stdout: modelID1 + "\n",
			},
		},
		{
			Name:    "success - create model with provider by ID",
			Command: []string{"model", "create", "claude-3-5-sonnet", "--model-provider", modelProviderID1, "--context-window", "200000"},
			SetupMocks: func(mockClient *api_client.MockClient) {
				setupModelCreateMock(mockClient, "claude-3-5-sonnet", modelProviderID1, 200000, modelID1, createdAt)
			},
			Expected: TestExpectation{
				Stdout: modelID1 + "\n",
			},
		},
		{
			Name:    "success - create model with short flags",
			Command: []string{"model", "create", "llama-3.1-8b", "-p", "ollama-local", "-w", "32768"},
			SetupMocks: func(mockClient *api_client.MockClient) {
				setupModelProviderLookupForCreateMock(mockClient, "ollama-local", modelProviderID1)
				setupModelCreateMock(mockClient, "llama-3.1-8b", modelProviderID1, 32768, modelID1, createdAt)
			},
			Expected: TestExpectation{
				Stdout: modelID1 + "\n",
			},
		},
		{
			Name:    "error - model provider not found by name",
			Command: []string{"model", "create", "test-model", "--model-provider", "nonexistent", "--context-window", "4096"},
			SetupMocks: func(mockClient *api_client.MockClient) {
				mockClient.ModelProvider.EXPECT().ListModelProviders(
					gomock.Any(),
					&connect.Request[v1.ListModelProvidersRequest]{
						Msg: &v1.ListModelProvidersRequest{},
					},
				).Return(&connect.Response[v1.ListModelProvidersResponse]{
					Msg: &v1.ListModelProvidersResponse{
						ModelProviders: []*v1.ModelProvider{},
					},
				}, nil)
			},
			Expected: TestExpectation{
				Error: "failed to resolve model provider nonexistent: model provider nonexistent not found",
			},
		},
		{
			Name:    "error - create model API failure",
			Command: []string{"model", "create", "test-model", "--model-provider", modelProviderID1, "--context-window", "4096"},
			SetupMocks: func(mockClient *api_client.MockClient) {
				mockClient.Model.EXPECT().CreateModel(
					gomock.Any(),
					&connect.Request[v1.CreateModelRequest]{
						Msg: &v1.CreateModelRequest{
							Name:            "test-model",
							ModelProviderId: modelProviderID1,
							ContextWindow:   4096,
						},
					},
				).Return(nil, connect.NewError(connect.CodeInternal, nil))
			},
			Expected: TestExpectation{
				Error: "failed to create model: internal",
			},
		},
		{
			Name:    "error - model provider lookup API failure",
			Command: []string{"model", "create", "test-model", "--model-provider", "openai-dev", "--context-window", "4096"},
			SetupMocks: func(mockClient *api_client.MockClient) {
				mockClient.ModelProvider.EXPECT().ListModelProviders(
					gomock.Any(),
					&connect.Request[v1.ListModelProvidersRequest]{
						Msg: &v1.ListModelProvidersRequest{},
					},
				).Return(nil, connect.NewError(connect.CodeInternal, nil))
			},
			Expected: TestExpectation{
				Error: "failed to resolve model provider openai-dev: failed to list model providers: internal",
			},
		},
	})
}

func setupModelCreateMock(mockClient *api_client.MockClient, name, modelProviderID string, contextWindow int64, modelID string, createdAt time.Time) {
	mockClient.Model.EXPECT().CreateModel(
		gomock.Any(),
		&connect.Request[v1.CreateModelRequest]{
			Msg: &v1.CreateModelRequest{
				Name:            name,
				ModelProviderId: modelProviderID,
				ContextWindow:   contextWindow,
			},
		},
	).Return(&connect.Response[v1.CreateModelResponse]{
		Msg: &v1.CreateModelResponse{
			Model: &v1.Model{
				Id:              modelID,
				Name:            name,
				ModelProviderId: modelProviderID,
				ContextWindow:   contextWindow,
				Enabled:         true,
				Capabilities:    []v1.ModelCapability{v1.ModelCapability_MODEL_CAPABILITY_IMAGE},
			},
		},
	}, nil)
}

func setupModelProviderLookupForCreateMock(mockClient *api_client.MockClient, providerName, providerID string) {
	mockClient.ModelProvider.EXPECT().ListModelProviders(
		gomock.Any(),
		&connect.Request[v1.ListModelProvidersRequest]{
			Msg: &v1.ListModelProvidersRequest{},
		},
	).Return(&connect.Response[v1.ListModelProvidersResponse]{
		Msg: &v1.ListModelProvidersResponse{
			ModelProviders: []*v1.ModelProvider{
				{
					Id:           providerID,
					Name:         providerName,
					ProviderType: v1.ModelProviderType_MODEL_PROVIDER_TYPE_OPENAI,
					Enabled:      true,
				},
			},
		},
	}, nil)
}
