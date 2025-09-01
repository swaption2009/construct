package model

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func SupportedXAIModels() []Model {
	return []Model{
		{
			ID:       uuid.MustParse("01980000-0001-7000-8000-000000000001"),
			Name:     "grok-code-fast-1",
			Provider: ProviderKindXAI,
			Capabilities: []Capability{
				CapabilityImage,
			},
			ContextWindow: 256000,
			Pricing: ModelPricing{
				Input:      0.20,
				Output:     1.50,
				CacheWrite: 0.0,
				CacheRead:  0.0,
			},
		},
		{
			ID:       uuid.MustParse("01980000-0002-7000-8000-000000000002"),
			Name:     "grok-4-0709",
			Provider: ProviderKindXAI,
			Capabilities: []Capability{
				CapabilityImage,
			},
			ContextWindow: 256000,
			Pricing: ModelPricing{
				Input:      3.00,
				Output:     15.00,
				CacheWrite: 0.0,
				CacheRead:  0.0,
			},
		},
		{
			ID:       uuid.MustParse("01980000-0003-7000-8000-000000000003"),
			Name:     "grok-3",
			Provider: ProviderKindXAI,
			Capabilities: []Capability{
				CapabilityImage,
			},
			ContextWindow: 131072,
			Pricing: ModelPricing{
				Input:      3.00,
				Output:     15.00,
				CacheWrite: 0.0,
				CacheRead:  0.0,
			},
		},
		{
			ID:       uuid.MustParse("01980000-0004-7000-8000-000000000004"),
			Name:     "grok-3-mini",
			Provider: ProviderKindXAI,
			Capabilities: []Capability{
				CapabilityImage,
			},
			ContextWindow: 131072,
			Pricing: ModelPricing{
				Input:      0.30,
				Output:     0.50,
				CacheWrite: 0.0,
				CacheRead:  0.0,
			},
		},
	}
}

func (p *AnthropicProvider) GetXAIModel(ctx context.Context, modelID uuid.UUID) (Model, error) {
	for _, model := range SupportedXAIModels() {
		if model.ID == modelID {
			return model, nil
		}
	}

	return Model{}, fmt.Errorf("model not supported")
}
