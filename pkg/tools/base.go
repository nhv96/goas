package tools

import (
	"encoding/json"
	"fmt"
)

type Tooler interface {
	String() string
	Run(any) (any, error)
}

type Base struct {
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

type Function struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Parameters  Parameters `json:"parameters"`
}

type Parameters struct {
	Type       string                    `json:"type"`
	Required   []string                  `json:"required"`
	Properties map[string]map[string]any `json:"properties"`
}

func (b Base) String() string {
	byt, err := json.Marshal(b)
	if err != nil {
		panic(err)
	}

	return string(byt)
}

func (b Base) Run(input any) any { return fmt.Sprint("implement me") }
