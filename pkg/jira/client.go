package jira

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	// RedHatJiraBaseURL is the base URL for Red Hat's Jira instance
	RedHatJiraBaseURL = "https://issues.redhat.com"
)

// Client represents a Jira API client
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	token      string // Personal Access Token
}

// ClientOption represents a configuration option for the client
type ClientOption func(*Client)

// WithToken sets the Personal Access Token for authentication
func WithToken(token string) ClientOption {
	return func(c *Client) {
		c.token = token
	}
}

// WithBaseURL sets a custom base URL for the Jira instance
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		c.BaseURL = baseURL
	}
}

// NewClient creates a new Jira client for issues.redhat.com
func NewClient(opts ...ClientOption) *Client {
	client := &Client{
		BaseURL: RedHatJiraBaseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	// Apply options
	for _, opt := range opts {
		opt(client)
	}

	return client
}

// GetTicket fetches a Jira ticket by its ID or key
func (c *Client) GetTicket(ticketID string) (*Ticket, error) {
	url := fmt.Sprintf("%s/rest/api/2/issue/%s", c.BaseURL, ticketID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	// Add authentication header if token is provided
	if c.token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, &TicketNotFoundError{TicketID: ticketID}
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(body),
		}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var jiraResp JiraResponse
	if err := json.Unmarshal(body, &jiraResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	ticket, err := c.parseTicket(&jiraResp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ticket: %w", err)
	}

	return ticket, nil
}

// parseTicket converts a JiraResponse to a Ticket struct
func (c *Client) parseTicket(resp *JiraResponse) (*Ticket, error) {
	fields := resp.Fields

	ticket := &Ticket{
		ID:  resp.ID,
		Key: resp.Key,
	}

	// Parse summary
	if summary, ok := fields["summary"].(string); ok {
		ticket.Summary = summary
	}

	// Parse description
	if description, ok := fields["description"].(string); ok {
		ticket.Description = description
	}

	// Parse status
	if statusField, ok := fields["status"].(map[string]interface{}); ok {
		ticket.Status = Status{
			ID:   getStringFromMap(statusField, "id"),
			Name: getStringFromMap(statusField, "name"),
		}
	}

	// Parse issue type
	if issueTypeField, ok := fields["issuetype"].(map[string]interface{}); ok {
		ticket.IssueType = IssueType{
			ID:   getStringFromMap(issueTypeField, "id"),
			Name: getStringFromMap(issueTypeField, "name"),
		}
	}

	// Parse priority
	if priorityField, ok := fields["priority"].(map[string]interface{}); ok {
		ticket.Priority = Priority{
			ID:   getStringFromMap(priorityField, "id"),
			Name: getStringFromMap(priorityField, "name"),
		}
	}

	// Parse assignee
	if assigneeField, ok := fields["assignee"].(map[string]interface{}); ok && assigneeField != nil {
		ticket.Assignee = &User{
			AccountID:    getStringFromMap(assigneeField, "accountId"),
			DisplayName:  getStringFromMap(assigneeField, "displayName"),
			EmailAddress: getStringFromMap(assigneeField, "emailAddress"),
		}
	}

	// Parse reporter
	if reporterField, ok := fields["reporter"].(map[string]interface{}); ok {
		ticket.Reporter = User{
			AccountID:    getStringFromMap(reporterField, "accountId"),
			DisplayName:  getStringFromMap(reporterField, "displayName"),
			EmailAddress: getStringFromMap(reporterField, "emailAddress"),
		}
	}

	// Parse timestamps
	if created, ok := fields["created"].(string); ok {
		if t, err := time.Parse("2006-01-02T15:04:05.000-0700", created); err == nil {
			ticket.Created = t
		}
	}

	if updated, ok := fields["updated"].(string); ok {
		if t, err := time.Parse("2006-01-02T15:04:05.000-0700", updated); err == nil {
			ticket.Updated = t
		}
	}

	// Parse labels
	if labelsField, ok := fields["labels"].([]interface{}); ok {
		for _, label := range labelsField {
			if labelStr, ok := label.(string); ok {
				ticket.Labels = append(ticket.Labels, labelStr)
			}
		}
	}

	// Parse components
	if componentsField, ok := fields["components"].([]interface{}); ok {
		for _, comp := range componentsField {
			if compMap, ok := comp.(map[string]interface{}); ok {
				component := Component{
					ID:          getStringFromMap(compMap, "id"),
					Name:        getStringFromMap(compMap, "name"),
					Description: getStringFromMap(compMap, "description"),
					Self:        getStringFromMap(compMap, "self"),
				}

				// Parse component lead
				if leadField, ok := compMap["lead"].(map[string]interface{}); ok && leadField != nil {
					component.Lead = &User{
						AccountID:    getStringFromMap(leadField, "accountId"),
						DisplayName:  getStringFromMap(leadField, "displayName"),
						EmailAddress: getStringFromMap(leadField, "emailAddress"),
					}
				}

				ticket.Components = append(ticket.Components, component)
			}
		}
	}

	// Parse project
	if projectField, ok := fields["project"].(map[string]interface{}); ok {
		ticket.Project = Project{
			ID:   getStringFromMap(projectField, "id"),
			Key:  getStringFromMap(projectField, "key"),
			Name: getStringFromMap(projectField, "name"),
		}
	}

	return ticket, nil
}

// TestAuthentication tests if the current authentication is valid
func (c *Client) TestAuthentication() error {
	if c.token == "" {
		return fmt.Errorf("no authentication token provided")
	}

	url := fmt.Sprintf("%s/rest/api/2/myself", c.BaseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("authentication failed: invalid token")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("authentication test failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetTicketSummary returns a formatted summary of the ticket including components
func (t *Ticket) GetTicketSummary() string {
	var summary strings.Builder

	summary.WriteString(fmt.Sprintf("Ticket: %s - %s\n", t.Key, t.Summary))
	summary.WriteString(fmt.Sprintf("Status: %s | Type: %s | Priority: %s\n",
		t.Status.Name, t.IssueType.Name, t.Priority.Name))

	if t.Assignee != nil {
		summary.WriteString(fmt.Sprintf("Assignee: %s\n", t.Assignee.DisplayName))
	} else {
		summary.WriteString("Assignee: Unassigned\n")
	}

	summary.WriteString(fmt.Sprintf("Reporter: %s\n", t.Reporter.DisplayName))

	if len(t.Components) > 0 {
		summary.WriteString("Components: ")
		for i, comp := range t.Components {
			if i > 0 {
				summary.WriteString(", ")
			}
			summary.WriteString(comp.Name)
			if comp.Lead != nil {
				summary.WriteString(fmt.Sprintf(" (Lead: %s)", comp.Lead.DisplayName))
			}
		}
		summary.WriteString("\n")
	}

	if len(t.Labels) > 0 {
		summary.WriteString(fmt.Sprintf("Labels: %s\n", strings.Join(t.Labels, ", ")))
	}

	return summary.String()
}

// getStringFromMap safely extracts a string value from a map
func getStringFromMap(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}