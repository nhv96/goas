package prompt

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// GenerateResponse used for /api/generate
type GenerateResponse struct {
	Model              string        `json:"model"`
	CreatedAt          time.Time     `json:"created_at"`
	Response           string        `json:"response"`
	Done               bool          `json:"done"`
	DoneReason         string        `json:"done_reason"`
	Context            []int         `json:"context"`
	TotalDuration      time.Duration `json:"total_duration"` // Parsed as nanoseconds
	LoadDuration       time.Duration `json:"load_duration"`  // Parsed as nanoseconds
	PromptEvalCount    int           `json:"prompt_eval_count"`
	PromptEvalDuration time.Duration `json:"prompt_eval_duration"` // Parsed as nanoseconds
	EvalCount          int           `json:"eval_count"`
	EvalDuration       time.Duration `json:"eval_duration"` // Parsed as nanoseconds
}

// DecodeResponse reads from an io.Reader and decodes the JSON into a ResponsePayload struct.
func DecodeGenerateResponse(r io.Reader) (*GenerateResponse, error) {
	var resp GenerateResponse
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&resp); err != nil {
		return nil, fmt.Errorf("failed to decode response JSON: %w", err)
	}
	return &resp, nil
}
