package tool

import (
	"fmt"

	"github.com/furisto/construct/backend/tool/codeact"
	"github.com/grafana/sobek"
)

const submitReportDescription = `
## Description
Communicates work results to the user, whether tasks are fully completed or partially done. This tool sends a structured message about progress, deliverables, and outcomes. After calling this tool, no additional work can be performed unless the user provides follow-up tasks.

## Parameters
- **summary** (string, required): A clear, concise summary of what was accomplished during the work session. This should highlight the key deliverables, changes made, or outcomes achieved, regardless of completion status.
- **completed** (boolean, required): Indicates whether the assigned task has been fully completed (true) or is still in progress/partial (false).
- **deliverables** (array, optional): An array of strings listing the specific files, features, or outputs that were created or modified during the work. Each item should be a brief description of a concrete deliverable.
- **next_steps** (string, optional): Optional suggestions for what should be done next, or any follow-up actions that might be beneficial to continue the work.

## Expected Output
The tool returns no direct output. The results are communicated to the user through the system.

## CRITICAL RULES
- **Final Action**: This tool MUST be used as the final step to submit your work
- **Session Termination**: After calling this tool, the work session ends and no additional work can be performed
- **Single Use Only**: This tool can only be called once per interpreter run - plan accordingly
- **Complete Information**: Include all relevant information in one call, as you cannot supplement it later

## Usage Examples

### Completed implementation
%[1]s
submit_report({
  summary: "Successfully implemented user authentication system with login, registration, and password reset functionality.",
  completed: true,
  deliverables: [
    "User model with email/password authentication",
    "Login and registration API endpoints", 
    "Password reset workflow",
    "Frontend authentication components",
    "Unit and integration tests"
  ],
  next_steps: "You can now test the authentication system by running 'npm start' and navigating to /login"
})
%[1]s

### Partial progress due to time constraints
%[1]s
submit_report({
  summary: "Made significant progress on the user authentication system. Completed backend API and database models, but frontend components still need implementation.",
  completed: false,
  deliverables: [
    "User authentication database schema",
    "Login and registration API endpoints",
    "Password hashing and validation logic",
    "Basic authentication middleware"
  ],
  next_steps: "Next: implement frontend login/registration forms and integrate with the backend APIs"
})
%[1]s

### Analysis results
%[1]s
submit_report({
  summary: "Completed security audit of the application and identified key vulnerabilities with recommendations.",
  completed: true,
  deliverables: [
    "Security audit report (security-audit.md)",
    "List of 12 identified vulnerabilities with severity ratings",
    "Prioritized remediation plan",
    "Updated security guidelines document"
  ]
})
%[1]s
`

const SubmitReportKey = "submit_report"

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

func (s *SubmitReportInput) Validate() error {
	if s.Summary == "" {
		return codeact.NewCustomError("summary is required", []string{
			"Please provide a clear summary of what was accomplished during the work session",
			"The summary should highlight key deliverables, changes made, or outcomes achieved",
		})
	}
	return nil
}

func NewSubmitReportTool() codeact.Tool {
	return codeact.NewOnDemandTool(
		ToolNameSubmitReport,
		fmt.Sprintf(submitReportDescription, "```"),
		submitReportHandler,
	)
}

func submitReportHandler(session *codeact.Session) func(call sobek.FunctionCall) sobek.Value {
	return func(call sobek.FunctionCall) sobek.Value {
		if len(call.Arguments) == 0 {
			session.Throw(codeact.NewCustomError("submit_report requires at least 1 argument", []string{
				"- **summary** (string, required): A clear, concise summary of what was accomplished",
				"- **completed** (boolean, required): Whether the task has been completed",
				"- **deliverables** (array, optional): List of specific files, features, or outputs created",
				"- **next_steps** (string, optional): Suggestions for follow-up actions",
			}))
		}

		obj := call.Argument(0).ToObject(session.VM)

		summary := ""
		if summaryVal := obj.Get("summary"); summaryVal != nil && summaryVal != sobek.Undefined() {
			summary = summaryVal.String()
		}

		var deliverables []string
		if deliverablesVal := obj.Get("deliverables"); deliverablesVal != nil && deliverablesVal != sobek.Undefined() {
			deliverablesObj := deliverablesVal.ToObject(session.VM)
			if deliverablesObj.ClassName() == "Array" {
				length := int(deliverablesObj.Get("length").ToInteger())
				for i := 0; i < length; i++ {
					item := deliverablesObj.Get(fmt.Sprintf("%d", i))
					if item != nil && item != sobek.Undefined() {
						deliverables = append(deliverables, item.String())
					}
				}
			}
		}

		nextSteps := ""
		if nextStepsVal := obj.Get("next_steps"); nextStepsVal != nil && nextStepsVal != sobek.Undefined() {
			nextSteps = nextStepsVal.String()
		}

		var completed bool
		if completedVal := obj.Get("completed"); completedVal != nil && completedVal != sobek.Undefined() {
			completed = completedVal.ToBoolean()
		}

		input := &SubmitReportInput{
			Summary:      summary,
			Completed:    completed,
			Deliverables: deliverables,
			NextSteps:    nextSteps,
		}

		result, err := submitReport(input)
		if err != nil {
			session.Throw(err)
		}

		fmt.Fprintln(session.System, "REPORT SUBMITTED")
		return session.VM.ToValue(result)
	}
}

func submitReport(input *SubmitReportInput) (*SubmitReportResult, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	return &SubmitReportResult{
		Summary:      input.Summary,
		Completed:    input.Completed,
		Deliverables: input.Deliverables,
		NextSteps:    input.NextSteps,
	}, nil
}
