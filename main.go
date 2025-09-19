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
	"github.com/briandowns/spinner"
	"github.com/fatih/color"
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
		color.Blue("🔐 Using Personal Access Token for authentication")
		jiraClient = jira.NewClient(
			jira.WithToken(token),
			jira.WithBaseURL(jiraBaseURL),
		)

		// Test authentication with spinner
		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		s.Suffix = " Testing authentication..."
		s.Start()
		if err := jiraClient.TestAuthentication(); err != nil {
			s.Stop()
			color.Red("❌ Authentication failed: %v", err)
			os.Exit(1)
		}
		s.Stop()
		color.Green("✅ Authentication successful")
	} else {
		color.Yellow("🌐 Using anonymous access (public tickets only)")
		jiraClient = jira.NewClient(jira.WithBaseURL(jiraBaseURL))
	}

	color.Cyan("🏠 Using Jira instance: %s", jiraBaseURL)

	// Fetch Jira ticket with spinner
	s := spinner.New(spinner.CharSets[35], 100*time.Millisecond)
	s.Suffix = fmt.Sprintf(" Fetching Jira ticket: %s", ticketID)
	s.Start()
	ticket, err := jiraClient.GetTicket(ticketID)
	s.Stop()
	if err != nil {
		color.Red("❌ Failed to fetch Jira ticket: %v", err)
		os.Exit(1)
	}

	color.Green("\n✅ Successfully fetched ticket")
	printTicketInfo(ticket)

	// Determine template path
	templateFilePath := templatePath
	if templateFilePath == "" {
		templateFilePath = prompt.GetDefaultTemplatePath()
	}

	// Load and render prompt template
	color.Cyan("\n📋 Loading prompt template: %s", templateFilePath)
	promptText, err := prompt.LoadAndRenderTemplate(templateFilePath, ticket)
	if err != nil {
		color.Red("❌ Failed to load prompt template: %v", err)
		os.Exit(1)
	}

	// Initialize Anthropic client
	color.Cyan("☁️  Using Google Cloud region: %s, project: %s", region, projectID)
	client := anthropic.NewClient(
		vertex.WithGoogleAuth(ctx, region, projectID),
	)

	// Generate implementation plan with spinner
	s = spinner.New(spinner.CharSets[11], 100*time.Millisecond)
	s.Suffix = " 🤖 Generating implementation plan with Claude..."
	s.Start()
	message, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		MaxTokens: 4096,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(promptText)),
		},
		Model: DefaultModel,
	})
	s.Stop()
	if err != nil {
		color.Red("❌ Failed to generate implementation plan: %v", err)
		os.Exit(1)
	}

	color.Green("\n✅ Implementation plan generated successfully!")
	printSeparator()
	color.HiMagenta("🚀 IMPLEMENTATION PLAN")
	printSeparator()

	var implementationPlan strings.Builder
	for i := range message.Content {
		content := message.Content[i].Text
		fmt.Printf("%+v\n", content)
		implementationPlan.WriteString(content)
		implementationPlan.WriteString("\n")
	}
	printSeparator()

	// Save implementation plan to file
	if err := saveImplementationPlan(ticketID, ticket, implementationPlan.String(), "implementation-plans"); err != nil {
		color.Yellow("⚠️  Warning: Failed to save implementation plan to file: %v", err)
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

	color.Green("\n💾 Implementation plan saved to: %s", filePath)
	return nil
}

// printSeparator prints a decorative separator
func printSeparator() {
	color.HiBlue("═══════════════════════════════════════════════════════════════")
}

// printTicketInfo prints formatted ticket information
func printTicketInfo(ticket *jira.Ticket) {
	fmt.Println()
	printSeparator()
	color.HiYellow("📋 TICKET INFORMATION")
	printSeparator()

	color.HiWhite("🎫 Ticket: ")
	color.Green("%s - %s", ticket.Key, ticket.Summary)

	color.HiWhite("📊 Status: ")
	color.Cyan("%s", ticket.Status.Name)

	color.HiWhite("🏷️  Type: ")
	color.Cyan("%s", ticket.IssueType.Name)

	color.HiWhite("⚡ Priority: ")
	color.Cyan("%s", ticket.Priority.Name)

	if ticket.Assignee != nil {
		color.HiWhite("👤 Assignee: ")
		color.Cyan("%s", ticket.Assignee.DisplayName)
	} else {
		color.HiWhite("👤 Assignee: ")
		color.Yellow("Unassigned")
	}

	color.HiWhite("📝 Reporter: ")
	color.Cyan("%s", ticket.Reporter.DisplayName)

	if len(ticket.Components) > 0 {
		color.HiWhite("🔧 Components: ")
		for i, comp := range ticket.Components {
			if i > 0 {
				fmt.Print(", ")
			}
			color.Cyan("%s", comp.Name)
			if comp.Lead != nil {
				color.White(" (Lead: %s)", comp.Lead.DisplayName)
			}
		}
		fmt.Println()
	} else {
		color.HiWhite("🔧 Components: ")
		color.Yellow("None")
	}

	if len(ticket.Labels) > 0 {
		color.HiWhite("🏷️  Labels: ")
		color.Cyan("%s", strings.Join(ticket.Labels, ", "))
	}

	color.HiWhite("📄 Description: ")
	color.White("%.200s...", ticket.Description)
	printSeparator()
}
