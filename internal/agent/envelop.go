package agent

// Envelop contains data from agent to the chat caller
type Envelop struct {
	From     string
	Content  string
	Thinking bool
}

func (r Envelop) IsThinking() bool {
	return r.Thinking
}

func (r Envelop) GetContent() string {
	return r.Content
}
