package cmd

import (
	"testing"
	"time"

	"connectrpc.com/connect"
	api_client "github.com/furisto/construct/api/go/client"
	v1 "github.com/furisto/construct/api/go/v1"
	"github.com/furisto/construct/shared/conv"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestMessageCreate(t *testing.T) {
	setup := &TestSetup{}

	messageID1 := uuid.New().String()
	taskID1 := uuid.New().String()
	agentID1 := uuid.New().String()
	createdAt := time.Now()

	setup.RunTests(t, []TestScenario{
		{
			Name:    "success - create message with task ID",
			Command: []string{"message", "create", taskID1, "Please implement a hello world function"},
			SetupMocks: func(mockClient *api_client.MockClient) {
				setupMessageCreateMock(mockClient, taskID1, "Please implement a hello world function", messageID1, taskID1, agentID1, createdAt)
			},
			Expected: TestExpectation{
				Stdout: conv.Ptr(messageID1 + "\n"),
			},
		},
		{
			Name:    "error - create message API failure",
			Command: []string{"message", "create", taskID1, "Test message"},
			SetupMocks: func(mockClient *api_client.MockClient) {
				mockClient.Message.EXPECT().CreateMessage(
					gomock.Any(),
					&connect.Request[v1.CreateMessageRequest]{
						Msg: &v1.CreateMessageRequest{
							TaskId:  taskID1,
							Content: "Test message",
						},
					},
				).Return(nil, connect.NewError(connect.CodeInternal, nil))
			},
			Expected: TestExpectation{
				Error: "failed to create message: internal",
			},
		},
	})
}

func setupMessageCreateMock(mockClient *api_client.MockClient, taskID, content, messageID, taskIDResponse, agentID string, createdAt time.Time) {
	mockClient.Message.EXPECT().CreateMessage(
		gomock.Any(),
		&connect.Request[v1.CreateMessageRequest]{
			Msg: &v1.CreateMessageRequest{
				TaskId:  taskID,
				Content: content,
			},
		},
	).Return(&connect.Response[v1.CreateMessageResponse]{
		Msg: &v1.CreateMessageResponse{
			Message: &v1.Message{
				Metadata: &v1.MessageMetadata{
					Id:        messageID,
					TaskId:    taskIDResponse,
					AgentId:   &agentID,
					Role:      v1.MessageRole_MESSAGE_ROLE_USER,
					CreatedAt: timestamppb.New(createdAt),
					UpdatedAt: timestamppb.New(createdAt),
				},
				Spec: &v1.MessageSpec{
					Content: []*v1.MessagePart{
						{
							Data: &v1.MessagePart_Text_{
								Text: &v1.MessagePart_Text{
									Content: content,
								},
							},
						},
					},
				},
			},
		},
	}, nil)
}
