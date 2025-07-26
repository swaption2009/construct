package communication

import (
	"github.com/furisto/construct/backend/tool/base"
)

type SubmitReportInput struct {
	Summary      string
	Completed    bool
	Deliverables []string
	NextSteps    string
}

type SubmitReportResult struct {
	Summary      string   `json:"summary"`
	Completed    bool     `json:"completed"`
	Deliverables []string `json:"deliverables"`
	NextSteps    string   `json:"next_steps"`
}

func SubmitReport(input *SubmitReportInput) (*SubmitReportResult, error) {
	if input.Summary == "" {
		return nil, base.NewCustomError("summary is required", []string{
			"Please provide a clear summary of what was accomplished during the work session",
			"The summary should highlight key deliverables, changes made, or outcomes achieved",
		})
	}

	return &SubmitReportResult{
		Summary:      input.Summary,
		Completed:    input.Completed,
		Deliverables: input.Deliverables,
		NextSteps:    input.NextSteps,
	}, nil
}
