package jira

import "fmt"

// TicketNotFoundError represents an error when a ticket is not found
type TicketNotFoundError struct {
	TicketID string
}

func (e *TicketNotFoundError) Error() string {
	return fmt.Sprintf("ticket %s not found", e.TicketID)
}

// APIError represents a general API error
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error %d: %s", e.StatusCode, e.Message)
}

// IsTicketNotFound checks if the error is a TicketNotFoundError
func IsTicketNotFound(err error) bool {
	_, ok := err.(*TicketNotFoundError)
	return ok
}