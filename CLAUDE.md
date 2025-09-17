# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based tool that fetches Jira tickets by ID, extracts relevant information, and generates implementation plans using Anthropic's Claude API via Google Cloud Vertex AI.

## Architecture

- **main.go**: Entry point with Anthropic SDK integration using Vertex AI authentication
- **pkg/jira/**: Package for Jira API integration with full ticket parsing
- **pkg/prompt/**: Package for prompt template loading and rendering
- **prompts/**: Directory containing prompt templates for AI generation
- **prompts/implementation-plan.md**: Default template for implementation plan prompts
- **implementation-plans/**: Directory where generated implementation plans are saved as markdown files

## Development Commands

### Build and Run
```bash
# Build the binary
go build

# With anonymous access (public tickets only)
./jig <TICKET_ID>

# With Personal Access Token via command line (long form)
./jig --token=<YOUR_PAT> <TICKET_ID>

# With Personal Access Token via command line (short form)
./jig -t <YOUR_PAT> <TICKET_ID>

# With Personal Access Token via environment variable
JIRA_TOKEN=<YOUR_PAT> ./jig <TICKET_ID>

# With custom Google Cloud settings
./jig --region=us-central1 --project-id=my-project <TICKET_ID>
./jig -r us-west1 -p my-project <TICKET_ID>

# With custom Jira instance
./jig --jira-base-url=https://my-jira.com <TICKET_ID>

# With custom prompt template
./jig --template=my-custom-prompt.md <TICKET_ID>

# Combined flags
./jig -t <YOUR_PAT> -r us-central1 -p my-project --jira-base-url=https://my-jira.com --template=custom.md <TICKET_ID>

# Show help
./jig --help
./jig -h

# Run directly with go run
go run main.go <TICKET_ID>
go run main.go --token=<YOUR_PAT> <TICKET_ID>
```

### Dependencies
```bash
go mod download
go mod tidy
```

### Build Binary
```bash
go build -o jig
```

## Key Configuration

- **Model**: Uses `claude-sonnet-4@20250514` as the default Anthropic model
- **Default Google Cloud Settings**:
  - Region: `us-east5` (configurable with `--region` flag)
  - Project ID: `itpc-gcp-hcm-pe-eng-claude` (configurable with `--project-id` flag)
- **Default Jira Instance**: `https://issues.redhat.com` (configurable with `--jira-base-url` flag)
- **Dependencies**:
  - Primary: `github.com/anthropics/anthropic-sdk-go v1.12.0`
  - CLI Framework: `github.com/spf13/cobra v1.10.1`

## Authentication

The Jira client supports two authentication modes:

1. **Anonymous Access**: For public tickets on issues.redhat.com
2. **Personal Access Token (PAT)**: For authenticated access to private tickets or other Jira instances

### Setting up Personal Access Token
1. Go to your Jira profile settings
2. Create a new Personal Access Token
3. Use it via:
   - Command line flag: `--token=<YOUR_PAT>` or `-t <YOUR_PAT>`
   - Environment variable: `JIRA_TOKEN=<YOUR_PAT>`

## Prompt Templates

The tool uses customizable prompt templates to generate implementation plans. Templates use Go's text/template syntax with ticket data.

### Default Template
- **Location**: `prompts/implementation-plan.md`
- **Variables Available**:
  - `{{.Summary}}` - Ticket title
  - `{{.Description}}` - Ticket description
  - `{{.Status}}` - Current status
  - `{{.IssueType}}` - Issue type (Bug, Story, etc.)
  - `{{.Priority}}` - Priority level
  - `{{.Components}}` - Components (if any)
  - `{{.Labels}}` - Labels (if any)
  - `{{.Assignee}}` - Assigned user
  - `{{.Reporter}}` - Reporter user

### Custom Templates
Create your own template using the `--template` flag:
```bash
./jig --template=my-custom-prompt.md RHEL-12345
```

## Output Files

The tool automatically saves all generated implementation plans to the `implementation-plans/` directory with the following format:

- **Filename**: `{TICKET_ID}_{TIMESTAMP}.md`
- **Content**:
  - Ticket metadata header
  - Complete implementation plan
  - Markdown formatted for easy reading

### Example Output File
```
implementation-plans/RHEL-12345_20240917_143052.md
```

### File Structure
```markdown
# Implementation Plan: Example Ticket Title

**Ticket ID:** RHEL-12345
**Generated:** 2024-09-17 14:30:52
**Status:** Open
**Type:** Story
**Priority:** Medium
**Assignee:** John Doe
**Reporter:** Jane Smith
**Components:** Security, Networking
**Labels:** urgent, p2

---

[Generated implementation plan content]
```

## Development Notes

- The project uses Go 1.24.7
- Uses Cobra CLI framework with comprehensive flag support
- Vertex AI integration requires proper Google Cloud authentication setup
- Supports custom Google Cloud regions and project IDs via CLI flags
- Supports custom Jira instances via `--jira-base-url` flag
- Configurable prompt templates via `--template` flag
- Template system uses Go's text/template with ticket data interpolation
- The prompts directory contains customizable templates
- Automatically saves implementation plans to `implementation-plans/` directory
- Generated files use format: `{TICKET_ID}_{TIMESTAMP}.md`
- The Jira client automatically tests authentication when a PAT is provided
- Built-in help system with examples and flag descriptions
- All configuration options have sensible defaults but can be overridden
