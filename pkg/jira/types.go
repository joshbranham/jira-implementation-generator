package jira

import "time"

// Ticket represents a Jira ticket with essential fields
type Ticket struct {
	ID          string    `json:"id"`
	Key         string    `json:"key"`
	Summary     string    `json:"summary"`
	Description string    `json:"description"`
	Status      Status    `json:"status"`
	IssueType   IssueType `json:"issuetype"`
	Priority    Priority  `json:"priority"`
	Assignee    *User     `json:"assignee"`
	Reporter    User      `json:"reporter"`
	Created     time.Time `json:"created"`
	Updated     time.Time `json:"updated"`
	Labels      []string  `json:"labels"`
	Components  []Component `json:"components"`
	Project     Project   `json:"project"`
}

// Status represents the status of a Jira ticket
type Status struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// IssueType represents the type of a Jira issue
type IssueType struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Priority represents the priority of a Jira issue
type Priority struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// User represents a Jira user
type User struct {
	AccountID   string `json:"accountId"`
	DisplayName string `json:"displayName"`
	EmailAddress string `json:"emailAddress"`
}

// Component represents a Jira project component
type Component struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Lead        *User  `json:"lead"`
	Self        string `json:"self"`
}

// Project represents a Jira project
type Project struct {
	ID   string `json:"id"`
	Key  string `json:"key"`
	Name string `json:"name"`
}

// JiraResponse wraps the API response from Jira
type JiraResponse struct {
	ID     string                 `json:"id"`
	Key    string                 `json:"key"`
	Fields map[string]interface{} `json:"fields"`
}