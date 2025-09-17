package prompt

import (
	"encoding/xml"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/joshbranham/jira-implementation-generator/pkg/jira"
)

// POMLDocument represents the root POML document structure
type POMLDocument struct {
	XMLName      xml.Name      `xml:"poml"`
	Role         string        `xml:"role"`
	Task         string        `xml:"task"`
	Context      POMLContext   `xml:"context"`
	Instructions []POMLRequirement `xml:"instructions>requirement"`
	OutputFormat POMLOutputFormat  `xml:"output-format"`
	Style        POMLStyle     `xml:"style"`
}

// POMLContext represents the context section
type POMLContext struct {
	Sections []POMLSection `xml:"section"`
}

// POMLSection represents a context section
type POMLSection struct {
	Name        string      `xml:"name,attr"`
	Title       string      `xml:"title"`
	Description string      `xml:"description"`
	Metadata    POMLMetadata `xml:"metadata"`
}

// POMLMetadata represents ticket metadata
type POMLMetadata struct {
	Status     string `xml:"status"`
	Type       string `xml:"type"`
	Priority   string `xml:"priority"`
	Assignee   string `xml:"assignee"`
	Reporter   string `xml:"reporter"`
	Components string `xml:"components"`
	Labels     string `xml:"labels"`
}

// POMLRequirement represents an instruction requirement
type POMLRequirement struct {
	Text string `xml:",chardata"`
}

// POMLOutputFormat represents the output format specification
type POMLOutputFormat struct {
	Sections []POMLOutputSection `xml:"section"`
}

// POMLOutputSection represents an output section
type POMLOutputSection struct {
	Name    string `xml:"name,attr"`
	Title   string `xml:"title"`
	Content string `xml:"content"`
}

// POMLStyle represents styling instructions
type POMLStyle struct {
	Formatting string `xml:"formatting"`
}

// LoadAndRenderPOMLTemplate loads a POML template and renders it with ticket data
func LoadAndRenderPOMLTemplate(templatePath string, ticket *jira.Ticket) (string, error) {
	// Read the POML template file
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to read POML template file %s: %w", templatePath, err)
	}

	// Create template data from ticket
	data := createTemplateData(ticket)

	// Parse template with Go templating first (for variable substitution)
	tmpl, err := template.New("poml").Parse(string(templateContent))
	if err != nil {
		return "", fmt.Errorf("failed to parse POML template: %w", err)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute POML template: %w", err)
	}

	// Parse the rendered POML XML
	var pomlDoc POMLDocument
	if err := xml.Unmarshal([]byte(buf.String()), &pomlDoc); err != nil {
		return "", fmt.Errorf("failed to parse POML XML: %w", err)
	}

	// Convert POML to plain text prompt
	prompt := convertPOMLToPrompt(&pomlDoc)
	return prompt, nil
}

// convertPOMLToPrompt converts a POML document to a plain text prompt
func convertPOMLToPrompt(doc *POMLDocument) string {
	var prompt strings.Builder

	// Add role
	if doc.Role != "" {
		prompt.WriteString(fmt.Sprintf("Role: %s\n\n", strings.TrimSpace(doc.Role)))
	}

	// Add task
	if doc.Task != "" {
		prompt.WriteString(fmt.Sprintf("Task: %s\n\n", strings.TrimSpace(doc.Task)))
	}

	// Add context
	if len(doc.Context.Sections) > 0 {
		prompt.WriteString("Context:\n")
		for _, section := range doc.Context.Sections {
			if section.Title != "" {
				prompt.WriteString(fmt.Sprintf("\nTicket: %s\n", section.Title))
			}
			if section.Description != "" {
				prompt.WriteString(fmt.Sprintf("Description: %s\n", section.Description))
			}

			// Add metadata
			if section.Metadata.Status != "" {
				prompt.WriteString(fmt.Sprintf("Status: %s\n", section.Metadata.Status))
			}
			if section.Metadata.Type != "" {
				prompt.WriteString(fmt.Sprintf("Type: %s\n", section.Metadata.Type))
			}
			if section.Metadata.Priority != "" {
				prompt.WriteString(fmt.Sprintf("Priority: %s\n", section.Metadata.Priority))
			}
			if section.Metadata.Assignee != "" {
				prompt.WriteString(fmt.Sprintf("Assignee: %s\n", section.Metadata.Assignee))
			}
			if section.Metadata.Reporter != "" {
				prompt.WriteString(fmt.Sprintf("Reporter: %s\n", section.Metadata.Reporter))
			}
			if section.Metadata.Components != "" {
				prompt.WriteString(fmt.Sprintf("Components: %s\n", section.Metadata.Components))
			}
			if section.Metadata.Labels != "" {
				prompt.WriteString(fmt.Sprintf("Labels: %s\n", section.Metadata.Labels))
			}
		}
		prompt.WriteString("\n")
	}

	// Add instructions
	if len(doc.Instructions) > 0 {
		prompt.WriteString("Instructions:\n")
		for _, req := range doc.Instructions {
			if strings.TrimSpace(req.Text) != "" {
				prompt.WriteString(fmt.Sprintf("- %s\n", strings.TrimSpace(req.Text)))
			}
		}
		prompt.WriteString("\n")
	}

	// Add output format
	if len(doc.OutputFormat.Sections) > 0 {
		prompt.WriteString("Please provide your response in the following format:\n\n")
		for _, section := range doc.OutputFormat.Sections {
			if section.Title != "" {
				prompt.WriteString(fmt.Sprintf("## %s\n", section.Title))
			}
			if section.Content != "" {
				prompt.WriteString(fmt.Sprintf("%s\n\n", strings.TrimSpace(section.Content)))
			}
		}
	}

	// Add style guidance
	if doc.Style.Formatting != "" {
		prompt.WriteString("Formatting Guidelines:\n")
		prompt.WriteString(fmt.Sprintf("%s\n", strings.TrimSpace(doc.Style.Formatting)))
	}

	return prompt.String()
}