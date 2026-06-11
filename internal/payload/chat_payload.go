package payload

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// ChatPayload is used for chat api
type ChatPayload struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	Think    bool          `json:"think"`
	Stream   bool          `json:"stream"`
}

// ChatMessage is an unit of message in a conversation with the model
type ChatMessage struct {
	Role    ChatRole `json:"role"`
	Content string   `json:"content"`
}

// ChatRole defines the roles of the sender of prompt
type ChatRole int

const (
	RoleSystem ChatRole = iota
	RoleUser
	RoleAssistant
	RoleTool
)

func (r ChatRole) String() string {
	switch r {
	case RoleSystem:
		return "system"
	case RoleUser:
		return "user"
	case RoleAssistant:
		return "assistant"
	case RoleTool:
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
		*r = RoleSystem
	case "user":
		*r = RoleUser
	case "assistant":
		*r = RoleAssistant
	case "tool":
		*r = RoleTool
	default:
		return fmt.Errorf("invalid chat role value: %s", s)
	}
	return nil
}

// CreateChatPayload expects input messages are included with history messages
func CreateChatPayload(model string, messages []ChatMessage, think, stream bool) (io.Reader, error) {
	p := &ChatPayload{
		Model:    model,
		Think:    think,
		Stream:   stream,
		Messages: messages,
	}

	payload, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}

	return bytes.NewBuffer(payload), nil
}

// ChatResponse is the response from model server
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

// ChatMessageResponse is the message content of each response
type ChatMessageResponse struct {
	Role     ChatRole `json:"role"` // always assistant
	Content  string   `json:"content"`
	Thinking string   `json:"thinking"`
}

// DecodeChatResponse decodes a single message from the model in non-stream mode
func DecodeChatResponse(msg io.Reader) (*ChatResponse, error) {
	var resp ChatResponse
	decoder := json.NewDecoder(msg)
	if err := decoder.Decode(&resp); err != nil {
		return nil, fmt.Errorf("failed to decode response JSON: %w", err)
	}
	return &resp, nil
}

// DecodeChatStreamResponse decodes a single message from the model in stream mode
func DecodeChatStreamResponse(msg io.Reader) (*ChatResponse, error) {
	var resp ChatResponse
	decoder := json.NewDecoder(msg)
	if err := decoder.Decode(&resp); err != nil {
		return nil, fmt.Errorf("failed to decode stream response JSON: %w", err)
	}
	return &resp, nil
}
