package stream

import (
	"context"
	"iter"
	"sync"

	v1 "github.com/furisto/construct/api/go/v1"
	"github.com/furisto/construct/backend/memory"
	"github.com/furisto/construct/backend/memory/message"
	"github.com/furisto/construct/backend/memory/schema/types"
	"github.com/google/uuid"
	"github.com/maypok86/otter"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type subscription struct {
	channel chan *v1.SubscribeResponse
}

func (s *subscription) Send(message *v1.SubscribeResponse) {
	s.channel <- message
}

func (s *subscription) Close() {
	close(s.channel)
}

type MessageBlockType string

const (
	MessageBlockTypeDelta    MessageBlockType = "delta"
	MessageBlockTypeComplete MessageBlockType = "complete"
)

type MessageBlock struct {
	Block    *types.MessageBlock
	Type     MessageBlockType
	Received map[string]bool
}

type EventHub struct {
	memory      *memory.Client
	messages    *otter.Cache[uuid.UUID, []*MessageBlock]
	subscribers map[uuid.UUID][]*subscription
	mu          sync.RWMutex
}

func NewMessageHub(db *memory.Client) (*EventHub, error) {
	messagesCache, err := otter.MustBuilder[uuid.UUID, []*MessageBlock](1000).Build()
	if err != nil {
		return nil, err
	}

	return &EventHub{
		memory:      db,
		messages:    &messagesCache,
		subscribers: make(map[uuid.UUID][]*subscription),
	}, nil
}

func (h *EventHub) Publish(taskID uuid.UUID, message *v1.SubscribeResponse) {
	h.mu.RLock()
	subscribers := make([]*subscription, len(h.subscribers[taskID]))
	copy(subscribers, h.subscribers[taskID])
	h.mu.RUnlock()

	for _, subscriber := range subscribers {
		subscriber.Send(message)
	}
}

func (h *EventHub) Subscribe(ctx context.Context, taskID uuid.UUID) iter.Seq2[*v1.SubscribeResponse, error] {
	subscription := &subscription{
		channel: make(chan *v1.SubscribeResponse, 64),
	}

	h.mu.Lock()
	h.subscribers[taskID] = append(h.subscribers[taskID], subscription)
	h.mu.Unlock()

	unsubscribe := func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		for i, s := range h.subscribers[taskID] {
			if s == subscription {
				h.subscribers[taskID] = append(h.subscribers[taskID][:i], h.subscribers[taskID][i+1:]...)
				break
			}
		}
	}

	return func(yield func(*v1.SubscribeResponse, error) bool) {
		defer unsubscribe()

		messages, err := h.memory.Message.Query().Where(message.TaskIDEQ(taskID)).Order(message.ByProcessedTime()).All(ctx)
		if err != nil {
			if !yield(nil, err) {
				return
			}
		}

		for _, m := range messages {
			protoMessage, err := ConvertMemoryMessageToProto(m)
			if err != nil {
				if !yield(nil, err) {
					return
				}
			}
			if !yield(&v1.SubscribeResponse{
				Message: protoMessage,
			}, nil) {
				return
			}
		}

		for {
			select {
			case <-ctx.Done():
				return
			case message := <-subscription.channel:
				if !yield(message, nil) {
					return
				}
			}
		}
	}
}

func ConvertMemoryMessageToProto(m *memory.Message) (*v1.Message, error) {
	var role v1.MessageRole
	switch m.Source {
	case types.MessageSourceUser:
		role = v1.MessageRole_MESSAGE_ROLE_USER
	case types.MessageSourceAssistant:
		role = v1.MessageRole_MESSAGE_ROLE_ASSISTANT
	default:
		role = v1.MessageRole_MESSAGE_ROLE_UNSPECIFIED
	}

	text := ""
	for _, block := range m.Content.Blocks {
		if block.Kind == types.MessageBlockKindText {
			text = block.Payload
			break
		}
	}

	messageUsage := &v1.MessageUsage{}
	if m.Usage != nil {
		messageUsage = &v1.MessageUsage{
			InputTokens:      m.Usage.InputTokens,
			OutputTokens:     m.Usage.OutputTokens,
			CacheWriteTokens: m.Usage.CacheWriteTokens,
		}
	}

	return &v1.Message{
		Metadata: &v1.MessageMetadata{
			Id:        m.ID.String(),
			CreatedAt: timestamppb.New(m.CreateTime),
			UpdatedAt: timestamppb.New(m.UpdateTime),
			TaskId:    m.TaskID.String(),
			AgentId:   func() *string { if m.AgentID != uuid.Nil { s := m.AgentID.String(); return &s }; return nil }(),
			ModelId:   func() *string { if m.ModelID != uuid.Nil { s := m.ModelID.String(); return &s }; return nil }(),
			Role:      role,
		},
		Spec: &v1.MessageSpec{
			Content: []*v1.MessagePart{
				{
					Data: &v1.MessagePart_Text_{
						Text: &v1.MessagePart_Text{
							Content: text,
						},
					},
				},
			},
		},
		Status: &v1.MessageStatus{
			Usage: messageUsage,
		},
	}, nil
}
