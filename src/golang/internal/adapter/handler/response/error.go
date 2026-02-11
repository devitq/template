package response

import (
	"time"
)

type ApiError struct {
	Code      string         `json:"code"`
	Message   string         `json:"message"`
	TraceId   string         `json:"traceId"`
	Timestamp time.Time      `json:"timestamp"`
	Path      string         `json:"path"`
	Details   map[string]any `json:"details,omitempty"`
}

type ValidationError struct {
	Code        string       `json:"code"`
	Message     string       `json:"message"`
	TraceId     string       `json:"traceId"`
	Timestamp   time.Time    `json:"timestamp"`
	Path        string       `json:"path"`
	FieldErrors []FieldError `json:"fieldErrors"`
}

type FieldError struct {
	Field         string `json:"field"`
	Issue         string `json:"issue"`
	RejectedValue any    `json:"rejectedValue,omitempty"`
}
