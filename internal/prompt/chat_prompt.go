package prompt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// ChatPrompt is the JSON data that be sent to Ollama server
type ChatPrompt struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	SharedPromptParams
}

// ChatMessage is an unit of message in a conversation with the model
type ChatMessage struct {
	Role    ChatRole `json:"role"`
	Content string   `json:"content"`
}

// ChatRole defines the roles of the sender of prompt
type ChatRole int

const (
	System ChatRole = iota
	User
	Assistant
	Tool
)

func (r ChatRole) String() string {
	switch r {
	case System:
		return "system"
	case User:
		return "user"
	case Assistant:
		return "assistant"
	case Tool:
		return "tool"
	default:
		panic("unimplemented chat role")
	}
}

// MarshalJSON overrides default marshal
func (r ChatRole) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.String())
}

// UnmarshalJSON overrides default marshal
func (r *ChatRole) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	switch s {
	case "system":
		*r = System
	case "user":
		*r = User
	case "assistant":
		*r = Assistant
	case "tool":
		*r = Tool
	default:
		return fmt.Errorf("invalid chat role value: %s", s)
	}
	return nil
}

// CreateChatPrompt returns io.Reader object.
func CreateChatPrompt(model, prompt string, chatHistory []ChatMessage, think, stream bool) (io.Reader, error) {
	p := &ChatPrompt{
		Model: model,
		SharedPromptParams: SharedPromptParams{
			Think:  think,
			Stream: stream,
		},
	}

	message := ChatMessage{
		Role:    User,
		Content: prompt,
	}

	p.Messages = chatHistory
	p.Messages = append(p.Messages, message)

	payload, err := json.MarshalIndent(p, "", "	")
	if err != nil {
		return nil, err
	}

	return bytes.NewBuffer(payload), nil
}

type ChatResponse struct {
	Model              string              `json:"model"`
	CreatedAt          time.Time           `json:"created_at"`
	Message            ChatMessageResponse `json:"message"`
	Done               bool                `json:"done"`
	DoneReason         string              `json:"done_reason"`
	TotalDuration      time.Duration       `json:"total_duration"` // Parsed as nanoseconds
	LoadDuration       time.Duration       `json:"load_duration"`  // Parsed as nanoseconds
	PromptEvalCount    int                 `json:"prompt_eval_count"`
	PromptEvalDuration time.Duration       `json:"prompt_eval_duration"` // Parsed as nanoseconds
	EvalCount          int                 `json:"eval_count"`
	EvalDuration       time.Duration       `json:"eval_duration"` // Parsed as nanoseconds
}

type ChatMessageResponse struct {
	Role     ChatRole `json:"role"` // always assistant
	Content  string   `json:"content"`
	Thinking string   `json:"thinking"`
}

func DecodeChatResponse(r io.Reader) (*ChatResponse, error) {
	var resp ChatResponse
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&resp); err != nil {
		return nil, fmt.Errorf("failed to decode response JSON: %w", err)
	}
	return &resp, nil
}
