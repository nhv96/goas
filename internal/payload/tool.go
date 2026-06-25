package payload

// Tool is the tool definition
type Tool struct {
	Type     string `json:"type"`
	Function struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Parameters  struct {
			Type       string                    `json:"type"`
			Required   []string                  `json:"required"`
			Properties map[string]map[string]any `json:"properties"`
		} `json:"parameters"`
	} `json:"function"`
}

// ToolCall is the function call to the tool
type ToolCall struct {
	Type     string `json:"type"`
	Function struct {
		Index     int            `json:"index"`
		Name      string         `json:"name"`
		Arguments map[string]any `json:"arguments"`
	} `json:"function"`
}
