package agent

// Reply is a way to pass data between agent and application
type Reply struct {
	From     string
	Content  string
	Thinking bool
}

func (r Reply) IsThinking() bool {
	return r.Thinking
}

func (r Reply) GetContent() string {
	return r.Content
}
