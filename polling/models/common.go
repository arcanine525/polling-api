package models

// ErrorResponse is the standard error envelope.
type ErrorResponse struct {
	Error string `json:"error" example:"something went wrong"`
}
