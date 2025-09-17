# Implementation Plan Prompt Template

Please create an implementation plan for this Jira ticket:

## Ticket Information

**Title:** {{.Summary}}
**Description:** {{.Description}}
**Status:** {{.Status}}
**Type:** {{.IssueType}}
**Priority:** {{.Priority}}
{{if .Components}}**Components:** {{.Components}}{{end}}
{{if .Labels}}**Labels:** {{.Labels}}{{end}}
**Assignee:** {{.Assignee}}
**Reporter:** {{.Reporter}}

## Required Implementation Plan

Please provide a detailed implementation plan including:

1. **Analysis of the requirements**
   - Break down the ticket requirements
   - Identify key functionality needed
   - Note any dependencies or prerequisites

2. **Proposed solution approach**
   - High-level architecture or design approach
   - Technology stack considerations
   - Integration points with existing systems

3. **Implementation steps**
   - Detailed step-by-step implementation plan
   - Logical sequence of development tasks
   - Estimated effort for each step

4. **Testing strategy**
   - Unit testing approach
   - Integration testing requirements
   - User acceptance testing criteria
   - Performance testing considerations (if applicable)

5. **Potential risks and mitigation**
   - Technical risks and how to address them
   - Timeline risks and contingency plans
   - Dependency risks and alternatives

6. **Any GitHub or GitLab projects that rely on the changes**
   - Downstream systems that may be affected
   - Required coordination with other teams
   - Documentation updates needed

## Output Format

Please structure your response clearly with headers and bullet points for easy readability.