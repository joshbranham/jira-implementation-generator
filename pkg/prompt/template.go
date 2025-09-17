package prompt

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/joshbranham/jira-implementation-generator/pkg/jira"
)

// TemplateData holds the data for prompt template rendering
type TemplateData struct {
	Summary    string
	Description string
	Status     string
	IssueType  string
	Priority   string
	Components string
	Labels     string
	Assignee   string
	Reporter   string
}

// LoadAndRenderTemplate loads a prompt template and renders it with ticket data
func LoadAndRenderTemplate(templatePath string, ticket *jira.Ticket) (string, error) {
	// Read the template file
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to read template file %s: %w", templatePath, err)
	}

	// Create template data from ticket
	data := createTemplateData(ticket)

	// Parse and execute template
	tmpl, err := template.New(filepath.Base(templatePath)).Parse(string(templateContent))
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// createTemplateData converts a Jira ticket to template data
func createTemplateData(ticket *jira.Ticket) TemplateData {
	data := TemplateData{
		Summary:     ticket.Summary,
		Description: ticket.Description,
		Status:      ticket.Status.Name,
		IssueType:   ticket.IssueType.Name,
		Priority:    ticket.Priority.Name,
		Reporter:    ticket.Reporter.DisplayName,
	}

	// Handle assignee (may be nil)
	if ticket.Assignee != nil {
		data.Assignee = ticket.Assignee.DisplayName
	} else {
		data.Assignee = "Unassigned"
	}

	// Handle components
	if len(ticket.Components) > 0 {
		var compNames []string
		for _, comp := range ticket.Components {
			if comp.Lead != nil {
				compNames = append(compNames, fmt.Sprintf("%s (Lead: %s)", comp.Name, comp.Lead.DisplayName))
			} else {
				compNames = append(compNames, comp.Name)
			}
		}
		data.Components = strings.Join(compNames, ", ")
	}

	// Handle labels
	if len(ticket.Labels) > 0 {
		data.Labels = strings.Join(ticket.Labels, ", ")
	}

	return data
}

// GetDefaultTemplatePath returns the default template path
func GetDefaultTemplatePath() string {
	return "prompts/implementation-plan.md"
}