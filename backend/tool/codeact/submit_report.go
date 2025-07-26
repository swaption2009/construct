package codeact

import (
	"fmt"

	"github.com/grafana/sobek"

	"github.com/furisto/construct/backend/tool/communication"
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

func NewSubmitReportTool() Tool {
	return NewOnDemandTool(
		"submit_report",
		fmt.Sprintf(submitReportDescription, "```"),
		submitReportInput,
		submitReportHandler,
	)
}

func submitReportInput(session *Session, args []sobek.Value) (any, error) {
	if len(args) == 0 {
		return nil, NewCustomError("submit_report requires at least 1 argument", []string{
			"- **summary** (string, required): A clear, concise summary of what was accomplished",
			"- **completed** (boolean, required): Whether the task has been completed",
			"- **deliverables** (array, optional): List of specific files, features, or outputs created",
			"- **next_steps** (string, optional): Suggestions for follow-up actions",
		})
	}

	obj := args[0].ToObject(session.VM)
	if obj == nil {
		return nil, nil
	}

	input := &communication.SubmitReportInput{}

	if summaryVal := obj.Get("summary"); summaryVal != nil && summaryVal != sobek.Undefined() {
		input.Summary = summaryVal.String()
	}

	if deliverablesVal := obj.Get("deliverables"); deliverablesVal != nil && deliverablesVal != sobek.Undefined() {
		deliverablesObj := deliverablesVal.ToObject(session.VM)
		if deliverablesObj != nil && deliverablesObj.ClassName() == "Array" {
			length := int(deliverablesObj.Get("length").ToInteger())
			for i := range length {
				item := deliverablesObj.Get(fmt.Sprintf("%d", i))
				if item != nil && item != sobek.Undefined() {
					input.Deliverables = append(input.Deliverables, item.String())
				}
			}
		}
	}

	if nextStepsVal := obj.Get("next_steps"); nextStepsVal != nil && nextStepsVal != sobek.Undefined() {
		input.NextSteps = nextStepsVal.String()
	}

	if completedVal := obj.Get("completed"); completedVal != nil && completedVal != sobek.Undefined() {
		input.Completed = completedVal.ToBoolean()
	}

	return input, nil
}

func submitReportHandler(session *Session) func(call sobek.FunctionCall) sobek.Value {
	return func(call sobek.FunctionCall) sobek.Value {
		rawInput, err := submitReportInput(session, call.Arguments)
		if err != nil {
			session.Throw(err)
		}
		input := rawInput.(*communication.SubmitReportInput)

		result, err := communication.SubmitReport(input)
		if err != nil {
			session.Throw(err)
		}

		SetValue(session, "result", result)
		fmt.Fprintln(session.System, "REPORT SUBMITTED")
		return session.VM.ToValue(result)
	}
}
