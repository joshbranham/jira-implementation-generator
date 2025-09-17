package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/vertex"
	"github.com/joshbranham/jira-implementation-generator/pkg/jira"
	"github.com/joshbranham/jira-implementation-generator/pkg/prompt"
	"github.com/spf13/cobra"
)

const DefaultModel = "claude-sonnet-4@20250514"
const DefaultProjectID = "itpc-gcp-hcm-pe-eng-claude"
const DefaultRegion = "us-east5"
const DefaultJiraBaseURL = "https://issues.redhat.com"

var (
	token        string
	region       string
	projectID    string
	jiraBaseURL  string
	templatePath string
)

var rootCmd = &cobra.Command{
	Use:   "jig <TICKET_ID>",
	Short: "Generate implementation plans for Jira tickets using Google Cloud Vertex AI",
	Long: `Jira Implementation Generator (jig) fetches Jira tickets from Jira
and generates detailed implementation plans using Google Cloud Vertex AI.

The tool supports both anonymous access for public tickets and authenticated
access using Personal Access Tokens for private tickets.`,
	Example: `  jig RHEL-12345
  jig --token=your_pat_here RHEL-12345
  jig --region=us-central1 --project-id=my-project RHEL-12345
  jig --jira-base-url=https://my-jira.com RHEL-12345`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runJiraGenerator(context.Background(), args[0])
	},
}

func init() {
	rootCmd.Flags().StringVarP(&token, "token", "t", "", "Jira Personal Access Token (can also be set via JIRA_TOKEN environment variable)")
	rootCmd.Flags().StringVarP(&region, "region", "r", DefaultRegion, "Google Cloud region for Vertex AI")
	rootCmd.Flags().StringVarP(&projectID, "project-id", "p", DefaultProjectID, "Google Cloud project ID for Vertex AI")
	rootCmd.Flags().StringVar(&jiraBaseURL, "jira-base-url", DefaultJiraBaseURL, "Base URL for Jira instance")
	rootCmd.Flags().StringVar(&templatePath, "template", "", "Path to custom prompt template file (defaults to prompts/implementation-plan.md)")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func runJiraGenerator(ctx context.Context, ticketID string) {
	// Get token from flag or environment variable
	if token == "" {
		token = os.Getenv("JIRA_TOKEN")
	}

	// Initialize Jira client with optional authentication and custom base URL
	var jiraClient *jira.Client
	if token != "" {
		fmt.Println("Using Personal Access Token for authentication")
		jiraClient = jira.NewClient(
			jira.WithToken(token),
			jira.WithBaseURL(jiraBaseURL),
		)

		// Test authentication
		fmt.Print("Testing authentication... ")
		if err := jiraClient.TestAuthentication(); err != nil {
			log.Fatalf("Authentication failed: %v", err)
		}
		fmt.Println("✓ Authentication successful")
	} else {
		fmt.Println("Using anonymous access (public tickets only)")
		jiraClient = jira.NewClient(jira.WithBaseURL(jiraBaseURL))
	}

	fmt.Printf("Using Jira instance: %s\n", jiraBaseURL)

	// Fetch Jira ticket
	fmt.Printf("Fetching Jira ticket: %s\n", ticketID)
	ticket, err := jiraClient.GetTicket(ticketID)
	if err != nil {
		log.Fatalf("Failed to fetch Jira ticket: %v", err)
	}

	fmt.Printf("Successfully fetched ticket: %s - %s\n", ticket.Key, ticket.Summary)
	fmt.Printf("Status: %s\n", ticket.Status.Name)
	fmt.Printf("Type: %s\n", ticket.IssueType.Name)
	fmt.Printf("Priority: %s\n", ticket.Priority.Name)

	if ticket.Assignee != nil {
		fmt.Printf("Assignee: %s\n", ticket.Assignee.DisplayName)
	} else {
		fmt.Printf("Assignee: Unassigned\n")
	}

	fmt.Printf("Reporter: %s\n", ticket.Reporter.DisplayName)

	if len(ticket.Components) > 0 {
		fmt.Printf("Components: ")
		for i, comp := range ticket.Components {
			if i > 0 {
				fmt.Printf(", ")
			}
			fmt.Printf("%s", comp.Name)
			if comp.Lead != nil {
				fmt.Printf(" (Lead: %s)", comp.Lead.DisplayName)
			}
		}
		fmt.Println()
	} else {
		fmt.Printf("Components: None\n")
	}

	if len(ticket.Labels) > 0 {
		fmt.Printf("Labels: %s\n", strings.Join(ticket.Labels, ", "))
	}

	fmt.Printf("Description: %.200s...\n", ticket.Description)

	// Determine template path
	templateFilePath := templatePath
	if templateFilePath == "" {
		templateFilePath = prompt.GetDefaultTemplatePath()
	}

	// Load and render prompt template
	fmt.Printf("Loading prompt template: %s\n", templateFilePath)
	promptText, err := prompt.LoadAndRenderTemplate(templateFilePath, ticket)
	if err != nil {
		log.Fatalf("Failed to load prompt template: %v", err)
	}

	// Initialize Anthropic client
	fmt.Printf("Using Google Cloud region: %s, project: %s\n", region, projectID)
	client := anthropic.NewClient(
		vertex.WithGoogleAuth(ctx, region, projectID),
	)

	// Generate implementation plan
	fmt.Println("\nGenerating implementation plan...")
	message, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		MaxTokens: 4096,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(promptText)),
		},
		Model: DefaultModel,
	})
	if err != nil {
		log.Fatalf("Failed to generate implementation plan: %v", err)
	}

	fmt.Println("\n=== IMPLEMENTATION PLAN ===")
	var implementationPlan strings.Builder
	for i := range message.Content {
		content := message.Content[i].Text
		fmt.Printf("%+v\n", content)
		implementationPlan.WriteString(content)
		implementationPlan.WriteString("\n")
	}

	// Save implementation plan to file
	if err := saveImplementationPlan(ticketID, ticket, implementationPlan.String(), "implementation-plans"); err != nil {
		log.Printf("Warning: Failed to save implementation plan to file: %v", err)
	}
}

// saveImplementationPlan saves the implementation plan to a markdown file
func saveImplementationPlan(ticketID string, ticket *jira.Ticket, plan string, dir string) error {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Generate filename with ticket ID and timestamp
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s.md", ticketID, timestamp)
	filePath := filepath.Join(dir, filename)

	// Create markdown content with metadata header
	var content strings.Builder
	content.WriteString(fmt.Sprintf("# Implementation Plan: %s\n\n", ticket.Summary))
	content.WriteString(fmt.Sprintf("**Ticket ID:** %s\n", ticketID))
	content.WriteString(fmt.Sprintf("**Generated:** %s\n", time.Now().Format("2006-01-02 15:04:05")))
	content.WriteString(fmt.Sprintf("**Status:** %s\n", ticket.Status.Name))
	content.WriteString(fmt.Sprintf("**Type:** %s\n", ticket.IssueType.Name))
	content.WriteString(fmt.Sprintf("**Priority:** %s\n", ticket.Priority.Name))

	if ticket.Assignee != nil {
		content.WriteString(fmt.Sprintf("**Assignee:** %s\n", ticket.Assignee.DisplayName))
	} else {
		content.WriteString("**Assignee:** Unassigned\n")
	}

	content.WriteString(fmt.Sprintf("**Reporter:** %s\n", ticket.Reporter.DisplayName))

	if len(ticket.Components) > 0 {
		var compNames []string
		for _, comp := range ticket.Components {
			compNames = append(compNames, comp.Name)
		}
		content.WriteString(fmt.Sprintf("**Components:** %s\n", strings.Join(compNames, ", ")))
	}

	if len(ticket.Labels) > 0 {
		content.WriteString(fmt.Sprintf("**Labels:** %s\n", strings.Join(ticket.Labels, ", ")))
	}

	content.WriteString("\n---\n\n")
	content.WriteString(plan)

	// Write to file
	if err := os.WriteFile(filePath, []byte(content.String()), 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", filePath, err)
	}

	fmt.Printf("\n✓ Implementation plan saved to: %s\n", filePath)
	return nil
}
